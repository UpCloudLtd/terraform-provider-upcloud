package loadbalancer

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
				Description:      "The name of the service must be unique within customer account.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateNameDiagFunc,
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
				Description: "Private network UUID where traffic will be routed. Must reside in load balancer zone.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"configured_status": {
				Description: "The service configured status indicates the service's current intended status. Managed by the customer.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     string(upcloud.LoadBalancerConfiguredStatusStarted),
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
			"dns_name": {
				Description: "DNS name of the load balancer",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)
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
	lb, err := svc.CreateLoadBalancer(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(lb.UUID)

	if diags = setLoadBalancerResourceData(d, lb); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "load balancer created", map[string]interface{}{"name": lb.Name, "uuid": lb.UUID})
	return diags
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	var err error
	svc := meta.(*service.ServiceContext)
	lb, err := svc.GetLoadBalancer(ctx, &request.GetLoadBalancerRequest{UUID: d.Id()})
	if err != nil {
		return handleResourceError(d.Get("name").(string), d, err)
	}

	if diags = setLoadBalancerResourceData(d, lb); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)

	lb, err := svc.ModifyLoadBalancer(ctx, &request.ModifyLoadBalancerRequest{
		UUID:             d.Id(),
		Name:             d.Get("name").(string),
		Plan:             d.Get("plan").(string),
		ConfiguredStatus: d.Get("configured_status").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if diags = setLoadBalancerResourceData(d, lb); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "load balancer updated", map[string]interface{}{"name": lb.Name, "uuid": lb.UUID})
	return diags
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.ServiceContext)
	if err := svc.DeleteLoadBalancer(ctx, &request.DeleteLoadBalancerRequest{UUID: d.Id()}); err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "load balancer deleted", map[string]interface{}{"name": d.Get("name").(string), "uuid": d.Id()})

	// Wait load balancer to shutdown before continuing so that e.g. network can be deleted (if needed)
	return diag.FromErr(waitLoadBalancerToShutdown(ctx, svc, d.Id()))
}

func setLoadBalancerResourceData(d *schema.ResourceData, lb *upcloud.LoadBalancer) (diags diag.Diagnostics) {
	if err := d.Set("name", lb.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("plan", lb.Plan); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("zone", lb.Zone); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("network", lb.NetworkUUID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("configured_status", lb.ConfiguredStatus); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("operational_state", lb.OperationalState); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("dns_name", lb.DNSName); err != nil {
		return diag.FromErr(err)
	}

	var frontends, backends, resolvers []string

	for _, f := range lb.Frontends {
		frontends = append(frontends, f.Name)
	}

	if err := d.Set("frontends", frontends); err != nil {
		return diag.FromErr(err)
	}

	for _, b := range lb.Backends {
		backends = append(backends, b.Name)
	}

	if err := d.Set("backends", backends); err != nil {
		return diag.FromErr(err)
	}

	for _, r := range lb.Resolvers {
		resolvers = append(resolvers, r.Name)
	}

	if err := d.Set("resolvers", resolvers); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func waitLoadBalancerToShutdown(ctx context.Context, svc *service.ServiceContext, id string) error {
	const maxRetries int = 100
	for i := 0; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			lb, err := svc.GetLoadBalancer(ctx, &request.GetLoadBalancerRequest{UUID: id})
			if err != nil {
				if svcErr, ok := err.(*upcloud.Problem); ok && svcErr.Status == http.StatusNotFound {
					return nil
				}
				return err
			}
			tflog.Info(ctx, "waiting load balancer to shutdown", map[string]interface{}{"name": lb.Name, "state": lb.OperationalState})
		}
		time.Sleep(5 * time.Second)
	}
	return errors.New("max retries reached while waiting for load balancer instance to shutdown")
}
