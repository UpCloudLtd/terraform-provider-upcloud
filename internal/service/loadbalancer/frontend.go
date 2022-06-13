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

func ResourceFrontend() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents load balancer frontend service",
		CreateContext: resourceFrontendCreate,
		ReadContext:   resourceFrontendRead,
		UpdateContext: resourceFrontendUpdate,
		DeleteContext: resourceFrontendDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"loadbalancer": {
				Description: "ID of the load balancer to which the frontend is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description:      "The name of the frontend must be unique within the load balancer service.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateNameDiagFunc,
			},
			"mode": {
				Description: "When load balancer operating in `tcp` mode it acts as a layer 4 proxy. In `http` mode it acts as a layer 7 proxy.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice([]string{
						string(upcloud.LoadBalancerModeHTTP),
						string(upcloud.LoadBalancerModeTCP),
					}, false),
				),
			},
			"port": {
				Description:      "Port to listen incoming requests",
				Type:             schema.TypeInt,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsPortNumber),
			},
			"default_backend_name": {
				Description: "The name of the default backend where traffic will be routed. Note, default backend can be overwritten in frontend rules.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"rules": {
				Description: "Set of frontend rule names",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tls_configs": {
				Description: "Set of TLS config names",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceFrontendCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)
	serviceID := d.Get("loadbalancer").(string)
	fe, err := svc.CreateLoadBalancerFrontend(ctx, &request.CreateLoadBalancerFrontendRequest{
		ServiceUUID: serviceID,
		Frontend: request.LoadBalancerFrontend{
			Name:           d.Get("name").(string),
			Mode:           upcloud.LoadBalancerMode(d.Get("mode").(string)),
			Port:           d.Get("port").(int),
			DefaultBackend: d.Get("default_backend_name").(string),
			Rules:          []request.LoadBalancerFrontendRule{},
			TLSConfigs:     []request.LoadBalancerFrontendTLSConfig{},
		},
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(marshalID(serviceID, fe.Name))

	if diags = setFrontendResourceData(d, fe); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] frontend '%s' created", fe.Name)
	return diags
}

func resourceFrontendRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)
	var serviceID, name string
	if err := unmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	fe, err := svc.GetLoadBalancerFrontend(ctx, &request.GetLoadBalancerFrontendRequest{
		ServiceUUID: serviceID,
		Name:        name,
	})

	if err != nil {
		return handleResourceError(d.Get("name").(string), d, err)
	}

	d.SetId(marshalID(serviceID, fe.Name))

	if err = d.Set("loadbalancer", serviceID); err != nil {
		return diag.FromErr(err)
	}

	if diags = setFrontendResourceData(d, fe); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceFrontendUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)
	var serviceID, name string
	if err := unmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	fe, err := svc.ModifyLoadBalancerFrontend(ctx, &request.ModifyLoadBalancerFrontendRequest{
		ServiceUUID: serviceID,
		Name:        name,
		Frontend: request.ModifyLoadBalancerFrontend{
			Name:           d.Get("name").(string),
			Mode:           upcloud.LoadBalancerMode(d.Get("mode").(string)),
			Port:           d.Get("port").(int),
			DefaultBackend: d.Get("default_backend_name").(string),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(marshalID(d.Get("loadbalancer").(string), fe.Name))

	if diags = setFrontendResourceData(d, fe); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] frontend '%s' updated", fe.Name)
	return diags
}

func resourceFrontendDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)
	var serviceID, name string
	if err := unmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] deleting frontend '%s'", d.Id())
	return diag.FromErr(svc.DeleteLoadBalancerFrontend(ctx, &request.DeleteLoadBalancerFrontendRequest{
		ServiceUUID: serviceID,
		Name:        name,
	}))
}

func setFrontendResourceData(d *schema.ResourceData, fe *upcloud.LoadBalancerFrontend) (diags diag.Diagnostics) {
	if err := d.Set("name", fe.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("mode", string(fe.Mode)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("port", fe.Port); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("default_backend_name", fe.DefaultBackend); err != nil {
		return diag.FromErr(err)
	}

	var tlsConfigs, rules []string

	for _, r := range fe.Rules {
		rules = append(rules, r.Name)
	}

	for _, t := range fe.TLSConfigs {
		tlsConfigs = append(tlsConfigs, t.Name)
	}

	if err := d.Set("rules", rules); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("tls_configs", tlsConfigs); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
