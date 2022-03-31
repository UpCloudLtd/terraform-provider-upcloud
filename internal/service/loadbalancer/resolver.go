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

func ResourceResolver() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents service's domain name resolver",
		CreateContext: resourceResolverCreate,
		ReadContext:   resourceResolverRead,
		UpdateContext: resourceResolverUpdate,
		DeleteContext: resourceResolverDelete,
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
				Description: "The name of the resolver must be unique within the service.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"nameservers": {
				Description: `List of nameserver IP addresses. Nameserver can reside in public internet or in customer private network. 
				Port is optional, if missing then default 53 will be used.`,
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 10,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"retries": {
				Description: "Number of retries on failure.",
				Type:        schema.TypeInt,
				Required:    true,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.IntBetween(1, 10),
				),
			},
			"timeout": {
				Description: "Timeout for the query in seconds.",
				Type:        schema.TypeInt,
				Required:    true,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.IntBetween(1, 60),
				),
			},
			"timeout_retry": {
				Description: "Timeout for the query retries in seconds.",
				Type:        schema.TypeInt,
				Required:    true,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.IntBetween(1, 60),
				),
			},
			"cache_valid": {
				Description: "Time in seconds to cache valid results.",
				Type:        schema.TypeInt,
				Required:    true,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.IntBetween(1, 86400),
				),
			},
			"cache_invalid": {
				Description: "Time in seconds to cache invalid results.",
				Type:        schema.TypeInt,
				Required:    true,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.IntBetween(1, 86400),
				),
			},
		},
	}
}

func resourceResolverCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	nameservers := make([]string, 0)
	if ns, ok := d.GetOk("nameservers"); ok {
		for _, s := range ns.([]interface{}) {
			nameservers = append(nameservers, s.(string))
		}
	}

	rs, err := svc.CreateLoadBalancerResolver(&request.CreateLoadBalancerResolverRequest{
		ServiceUUID: d.Get("loadbalancer").(string),
		Resolver: request.LoadBalancerResolver{
			Name:         d.Get("name").(string),
			Nameservers:  nameservers,
			Retries:      d.Get("retries").(int),
			Timeout:      d.Get("timeout").(int),
			TimeoutRetry: d.Get("timeout_retry").(int),
			CacheValid:   d.Get("cache_valid").(int),
			CacheInvalid: d.Get("cache_invalid").(int),
		},
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(marshalID(d.Get("loadbalancer").(string), rs.Name))

	if diags = setResolverResourceData(d, rs); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] resolver '%s' created", rs.Name)
	return diags
}

func resourceResolverRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, name string
	if err := unmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	rs, err := svc.GetLoadBalancerResolver(&request.GetLoadBalancerResolverRequest{
		ServiceUUID: serviceID,
		Name:        name,
	})

	if err != nil {
		return handleResourceError(d.Get("name").(string), d, err)
	}

	d.SetId(marshalID(d.Get("loadbalancer").(string), rs.Name))

	if diags = setResolverResourceData(d, rs); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceResolverUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	nameservers := make([]string, 0)
	if ns, ok := d.GetOk("nameservers"); ok {
		for _, s := range ns.([]interface{}) {
			nameservers = append(nameservers, s.(string))
		}
	}

	var serviceID, name string
	if err := unmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	rs, err := svc.ModifyLoadBalancerResolver(&request.ModifyLoadBalancerRevolverRequest{
		ServiceUUID: serviceID,
		Name:        name,
		Resolver: request.LoadBalancerResolver{
			Name:         d.Get("name").(string),
			Nameservers:  nameservers,
			Retries:      d.Get("retries").(int),
			Timeout:      d.Get("timeout").(int),
			TimeoutRetry: d.Get("timeout_retry").(int),
			CacheValid:   d.Get("cache_valid").(int),
			CacheInvalid: d.Get("cache_invalid").(int),
		},
	})

	d.SetId(marshalID(d.Get("loadbalancer").(string), rs.Name))

	if err != nil {
		return diag.FromErr(err)
	}

	if diags = setResolverResourceData(d, rs); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] resolver '%s' updated", rs.Name)
	return diags
}

func resourceResolverDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)
	var serviceID, name string
	if err := unmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] deleting resolver '%s'", d.Id())
	return diag.FromErr(
		svc.DeleteLoadBalancerResolver(&request.DeleteLoadBalancerResolverRequest{
			ServiceUUID: serviceID,
			Name:        name,
		}),
	)
}

func setResolverResourceData(d *schema.ResourceData, rs *upcloud.LoadBalancerResolver) (diags diag.Diagnostics) {
	if err := d.Set("name", rs.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("nameservers", rs.Nameservers); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("retries", rs.Retries); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("timeout", rs.Timeout); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("timeout_retry", rs.TimeoutRetry); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cache_valid", rs.CacheValid); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cache_invalid", rs.CacheInvalid); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
