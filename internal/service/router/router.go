package router

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const routerNotFoundErrorCode string = "ROUTER_NOT_FOUND"

func ResourceRouter() *schema.Resource {
	return &schema.Resource{
		Description: `This resource represents a generated UpCloud router resource. 
		Routers can be used to connect multiple Private Networks. 
		UpCloud Servers on any attached network can communicate directly with each other.`,
		CreateContext: resourceRouterCreate,
		ReadContext:   resourceRouterRead,
		UpdateContext: resourceRouterUpdate,
		DeleteContext: resourceRouterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the router",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "The type of router",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"attached_networks": {
				Description: "A collection of UUID representing networks attached to this router",
				Computed:    true,
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceRouterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.ServiceContext)

	var diags diag.Diagnostics

	opts := &request.CreateRouterRequest{
		Name: d.Get("name").(string),
	}

	router, err := client.CreateRouter(ctx, opts)

	if err != nil {
		return diag.FromErr(err)
	}

	attachedNetworks := make([]string, len(router.AttachedNetworks))

	for _, network := range router.AttachedNetworks {
		attachedNetworks = append(attachedNetworks, network.NetworkUUID)
	}

	if err := d.Set("attached_networks", attachedNetworks); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(router.UUID)

	return diags
}

func resourceRouterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.ServiceContext)

	var diags diag.Diagnostics

	opts := &request.GetRouterDetailsRequest{
		UUID: d.Id(),
	}

	router, err := client.GetRouterDetails(ctx, opts)

	if err != nil {
		if svcErr, ok := err.(*upcloud.Error); ok && svcErr.ErrorCode == routerNotFoundErrorCode {
			diags = append(diags, utils.DiagBindingRemovedWarningFromUpcloudErr(svcErr, d.Get("name").(string)))
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("name", router.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", router.Type); err != nil {
		return diag.FromErr(err)
	}

	attachedNetworks := make([]string, len(router.AttachedNetworks))
	for i, network := range router.AttachedNetworks {
		attachedNetworks[i] = network.NetworkUUID
	}

	if err := d.Set("attached_networks", attachedNetworks); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceRouterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.ServiceContext)

	opts := &request.ModifyRouterRequest{
		UUID: d.Id(),
	}

	if v, ok := d.GetOk("name"); ok {
		opts.Name = v.(string)
	}

	_, err := client.ModifyRouter(ctx, opts)

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRouterRead(ctx, d, meta)
}

func resourceRouterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.ServiceContext)
	var diags diag.Diagnostics

	router, err := client.GetRouterDetails(ctx, &request.GetRouterDetailsRequest{
		UUID: d.Id(),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if len(router.AttachedNetworks) > 0 {
		for _, network := range router.AttachedNetworks {
			err := client.DetachNetworkRouter(ctx, &request.DetachNetworkRouterRequest{
				NetworkUUID: network.NetworkUUID,
			})
			if err != nil {
				return diag.FromErr(fmt.Errorf("cannot detach from network %v: %w", network.NetworkUUID, err))
			}
		}
	}
	err = client.DeleteRouter(ctx, &request.DeleteRouterRequest{
		UUID: d.Id(),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
