package upcloud

import (
	"context"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUpCloudRouter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUpCloudRouterCreate,
		ReadContext:   resourceUpCloudRouterRead,
		UpdateContext: resourceUpCloudRouterUpdate,
		DeleteContext: resourceUpCloudRouterDelete,
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

func resourceUpCloudRouterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*service.Service)

	var diags diag.Diagnostics

	opts := &request.CreateRouterRequest{
		Name: d.Get("name").(string),
	}

	router, err := client.CreateRouter(opts)

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

func resourceUpCloudRouterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*service.Service)

	var diags diag.Diagnostics

	opts := &request.GetRouterDetailsRequest{
		UUID: d.Id(),
	}

	router, err := client.GetRouterDetails(opts)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", router.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", router.Type); err != nil {
		return diag.FromErr(err)
	}

	attachedNetworks := make([]string, len(router.AttachedNetworks))

	for _, network := range router.AttachedNetworks {
		attachedNetworks = append(attachedNetworks, network.NetworkUUID)
	}

	if err := d.Set("attached_networks", attachedNetworks); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceUpCloudRouterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*service.Service)

	opts := &request.ModifyRouterRequest{
		UUID: d.Id(),
	}

	if v, ok := d.GetOk("name"); ok {
		opts.Name = v.(string)
	}

	_, err := client.ModifyRouter(opts)

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUpCloudRouterRead(ctx, d, meta)
}

func resourceUpCloudRouterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*service.Service)

	var diags diag.Diagnostics

	opts := &request.DeleteRouterRequest{
		UUID: d.Id(),
	}

	err := client.DeleteRouter(opts)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
