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
				Description: "ID of the load balancer to which the backend is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description:      "The name of the backend must be unique within the load balancer service.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateNameDiagFunc,
			},
			"resolver_name": {
				Description: "Domain Name Resolver used with dynamic type members.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"members": {
				Description: "Backend members receive traffic dispatched from the frontends",
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
	serviceID := d.Get("loadbalancer").(string)

	be, err := svc.CreateLoadBalancerBackend(&request.CreateLoadBalancerBackendRequest{
		ServiceUUID: serviceID,
		Backend: request.LoadBalancerBackend{
			Name:     d.Get("name").(string),
			Resolver: d.Get("resolver_name").(string),
			Members:  []request.LoadBalancerBackendMember{},
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(marshalID(serviceID, be.Name))

	if diags = setBackendResourceData(d, be); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] backend '%s' created", be.Name)
	return diags
}

func resourceBackendRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, name string
	if err := unmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	be, err := svc.GetLoadBalancerBackend(&request.GetLoadBalancerBackendRequest{
		ServiceUUID: serviceID,
		Name:        name,
	})

	if err != nil {
		return handleResourceError(d.Get("name").(string), d, err)
	}

	d.SetId(marshalID(serviceID, be.Name))

	if err = d.Set("loadbalancer", serviceID); err != nil {
		return diag.FromErr(err)
	}

	if diags = setBackendResourceData(d, be); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceBackendUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, name string
	if err := unmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}

	be, err := svc.ModifyLoadBalancerBackend(&request.ModifyLoadBalancerBackendRequest{
		ServiceUUID: serviceID,
		Name:        name,
		Backend: request.ModifyLoadBalancerBackend{
			Name:     d.Get("name").(string),
			Resolver: upcloud.StringPtr(d.Get("resolver_name").(string)),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(marshalID(d.Get("loadbalancer").(string), be.Name))

	if diags = setBackendResourceData(d, be); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] backend '%s' updated", be.Name)
	return diags
}

func resourceBackendDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, name string
	if err := unmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] deleting backend '%s'", d.Id())
	return diag.FromErr(svc.DeleteLoadBalancerBackend(&request.DeleteLoadBalancerBackendRequest{
		ServiceUUID: serviceID,
		Name:        name,
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
