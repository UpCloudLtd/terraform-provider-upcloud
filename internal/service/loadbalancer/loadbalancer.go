package loadbalancer

import (
	"context"
	"log"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents load balancer service",
		CreateContext: resourceLoadBalancerCreate,
		ReadContext:   resourceLoadBalancerRead,
		UpdateContext: resourceLoadBalancerUpdate,
		DeleteContext: resourceLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the service must be unique within customer account.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"plan": {
				Description: "Plan which the service will have",
				Type:        schema.TypeString,
				Required:    true,
			},
			"zone": {
				Description: "Zone in which the service will be hosted, e.g. `fi-hel1`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"network": {
				Description: "Private network UUID where traffic will be routed. Must reside in loadbalancer zone.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
			},
			"configured_status": {
				Description: "The service configured status indicates the service's current intended status. Managed by the customer.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice([]string{
						string(upcloud.LoadBalancerConfiguredStatusStarted),
						string(upcloud.LoadBalancerConfiguredStatusStarted),
					}, false),
				),
			},
			"frontends": {
				Description: "Frontends receive the traffic before dispatching it to the backends.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"backends": {
				Description: "Backends are groups of customer servers whose traffic should be balanced.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"resolvers": {
				Description: "Domain Name Resolvers must be configured in case of customer uses dynamic type members",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"operational_state": {
				Description: "The service operational state indicates the service's current operational, effective state. Managed by the system.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	req := &request.CreateLoadBalancerRequest{
		Name:             d.Get("name").(string),
		Plan:             d.Get("plan").(string),
		Zone:             d.Get("zone").(string),
		NetworkUUID:      d.Get("network").(string),
		ConfiguredStatus: upcloud.LoadBalancerConfiguredStatus(d.Get("configured_status").(string)),
		Frontends:        []request.LoadBalancerFrontend{},
		Backends:         []request.LoadBalancerBackend{},
		Resolvers:        []request.LoadBalancerResolver{},
	}
	lb, err := svc.CreateLoadBalancer(req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(lb.UUID)

	log.Printf("[INFO] load balancer '%s' created", lb.Name)
	return resourceLoadBalancerRead(ctx, d, meta)
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	var err error
	svc := meta.(*service.Service)
	lb, err := svc.GetLoadBalancer(&request.GetLoadBalancerRequest{UUID: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("name", lb.Name); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("plan", lb.Plan); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("zone", lb.Zone); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("network", lb.NetworkUUID); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("configured_status", lb.ConfiguredStatus); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("operational_state", lb.OperationalState); err != nil {
		return diag.FromErr(err)
	}

	var frontends, backends, resolvers []string

	for _, f := range lb.Frontends {
		frontends = append(frontends, f.Name)
	}

	if err = d.Set("frontends", frontends); err != nil {
		return diag.FromErr(err)
	}

	for _, b := range lb.Backends {
		backends = append(backends, b.Name)
	}

	if err = d.Set("backends", backends); err != nil {
		return diag.FromErr(err)
	}

	for _, r := range lb.Resolvers {
		resolvers = append(resolvers, r.Name)
	}

	if err = d.Set("resolvers", resolvers); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	lb, err := svc.ModifyLoadBalancer(&request.ModifyLoadBalancerRequest{
		UUID:             d.Id(),
		Name:             d.Get("name").(string),
		Plan:             d.Get("plan").(string),
		ConfiguredStatus: d.Get("configured_status").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] load balancer '%s' updated", lb.Name)
	return diags
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)
	return diag.FromErr(
		svc.DeleteLoadBalancer(&request.DeleteLoadBalancerRequest{UUID: d.Id()}),
	)
}
