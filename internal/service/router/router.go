package router

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceRouter() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
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
			"static_route": {
				Description: "A collection of static routes for this router",
				Optional:    true,
				Type:        schema.TypeSet,
				Elem: &schema.Resource{
					EnableLegacyTypeSystemApplyErrors: true,
					EnableLegacyTypeSystemPlanErrors:  true,
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "Name or description of the route.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"nexthop": {
							Description:  "Next hop address. NOTE: For static route to be active the next hop has to be an address of a reachable running Cloud Server in one of the Private Networks attached to the router.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.Any(validation.IsIPv4Address, validation.IsIPv6Address),
						},
						"route": {
							Description:  "Destination prefix of the route.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.Any(validation.IsCIDR),
						},
					},
				},
			},
		},
	}
}

func resourceRouterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	client := meta.(*service.Service)

	req := &request.CreateRouterRequest{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("static_route"); ok {
		for _, staticRoute := range v.(*schema.Set).List() {
			staticRouteData := staticRoute.(map[string]interface{})

			r := upcloud.StaticRoute{
				Name:    staticRouteData["name"].(string),
				Nexthop: staticRouteData["nexthop"].(string),
				Route:   staticRouteData["route"].(string),
			}

			req.StaticRoutes = append(req.StaticRoutes, r)
		}
	}

	router, err := client.CreateRouter(ctx, req)
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

	var staticRoutes []map[string]interface{}
	for _, staticRoute := range router.StaticRoutes {
		staticRoutes = append(staticRoutes, map[string]interface{}{
			"name":    staticRoute.Name,
			"nexthop": staticRoute.Nexthop,
			"route":   staticRoute.Route,
		})
	}

	if err := d.Set("static_route", staticRoutes); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(router.UUID)

	return diags
}

func resourceRouterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	client := meta.(*service.Service)

	opts := &request.GetRouterDetailsRequest{
		UUID: d.Id(),
	}

	router, err := client.GetRouterDetails(ctx, opts)
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
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

	var staticRoutes []map[string]interface{}
	for _, staticRoute := range router.StaticRoutes {
		staticRoutes = append(staticRoutes, map[string]interface{}{
			"name":    staticRoute.Name,
			"nexthop": staticRoute.Nexthop,
			"route":   staticRoute.Route,
		})
	}

	if err := d.Set("static_route", staticRoutes); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceRouterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	req := &request.ModifyRouterRequest{
		UUID: d.Id(),
	}

	if v, ok := d.GetOk("name"); ok {
		req.Name = v.(string)
	}

	var staticRoutes []upcloud.StaticRoute

	if v, ok := d.GetOk("static_route"); ok {
		for _, staticRoute := range v.(*schema.Set).List() {
			staticRouteData := staticRoute.(map[string]interface{})

			staticRoutes = append(staticRoutes, upcloud.StaticRoute{
				Name:    staticRouteData["name"].(string),
				Nexthop: staticRouteData["nexthop"].(string),
				Route:   staticRouteData["route"].(string),
			})
		}
	}

	req.StaticRoutes = &staticRoutes

	_, err := client.ModifyRouter(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRouterRead(ctx, d, meta)
}

func resourceRouterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	client := meta.(*service.Service)

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
