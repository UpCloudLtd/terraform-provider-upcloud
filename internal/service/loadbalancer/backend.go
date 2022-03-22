package loadbalancer

import (
	"context"
	"log"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceBackend() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents load balancer backend service",
		CreateContext: resourceBackendCreate,
		ReadContext:   resourceBackendRead,
		UpdateContext: resourceBackendUpdate,
		DeleteContext: resourceBackendDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"loadbalancer": {
				Description: "ID of the load balancer to which the resolver is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the backend must be unique within the load balancer service.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"resolver_name": {
				Description: "Domain Name Resolver used with dynamic type members.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"members": {
				Description: "Frontends receive the traffic before dispatching it to the backends.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceBackendCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	be, err := svc.CreateLoadBalancerBackend(&request.CreateLoadBalancerBackendRequest{
		ServiceUUID: d.Get("loadbalancer").(string),
		Backend: request.LoadBalancerBackend{
			Name:     d.Get("name").(string),
			Resolver: d.Get("resolver_name").(string),
			Members:  []request.LoadBalancerBackendMember{},
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if diags = setBackendResourceData(d, be); len(diags) > 0 {
		return diags
	}

	d.SetId(be.Name)

	log.Printf("[INFO] backend '%s' created", be.Name)
	return diags
}

func resourceBackendRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	be, err := svc.GetLoadBalancerBackend(&request.GetLoadBalancerBackendRequest{
		ServiceUUID: d.Get("loadbalancer").(string),
		Name:        d.Id(),
	})

	if err != nil {
		return handleResourceError(d.Get("name").(string), d, err)
	}

	if diags = setBackendResourceData(d, be); len(diags) > 0 {
		return diags
	}

	d.SetId(be.Name)

	return diags
}

func resourceBackendUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	be, err := svc.ModifyLoadBalancerBackend(&request.ModifyLoadBalancerBackendRequest{
		ServiceUUID: d.Get("loadbalancer").(string),
		Name:        d.Id(),
		Backend: request.ModifyLoadBalancerBackend{
			Name:     d.Get("name").(string),
			Resolver: d.Get("resolver_name").(string),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(be.Name)

	if diags = setBackendResourceData(d, be); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] backend '%s' updated", be.Name)
	return diags
}

func resourceBackendDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	log.Printf("[INFO] deleting backend '%s'", d.Id())
	return diag.FromErr(svc.DeleteLoadBalancerBackend(&request.DeleteLoadBalancerBackendRequest{
		ServiceUUID: d.Get("loadbalancer").(string),
		Name:        d.Id(),
	}))
}

func setBackendResourceData(d *schema.ResourceData, be *upcloud.LoadBalancerBackend) (diags diag.Diagnostics) {
	if err := d.Set("name", be.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("resolver_name", be.Resolver); err != nil {
		return diag.FromErr(err)
	}

	var members []string
	for _, m := range be.Members {
		members = append(members, m.Name)
	}

	if err := d.Set("members", members); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
