package loadbalancer

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
			"properties": {
				Description: "Frontend properties. Properties can set back to defaults by defining empty `properties {}` block.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: schemaFrontendProperties(),
				},
			},
			"networks": {
				Description: "Networks that frontend will be listening. Networks are required if load balancer has `networks` defined. " +
					"This field will be required when deprecated field `network` is removed from load balancer resource.",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "Name of the load balancer network",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func resourceFrontendCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
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
			Properties:     frontendPropertiesFromResourceData(d),
			Networks:       loadBalancerFrontendNetworksFromResourceData(d),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(serviceID, fe.Name))

	if diags = setFrontendResourceData(d, fe); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "frontend created", map[string]interface{}{"name": fe.Name, "service_uuid": serviceID})
	return diags
}

func resourceFrontendRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	fe, err := svc.GetLoadBalancerFrontend(ctx, &request.GetLoadBalancerFrontendRequest{
		ServiceUUID: serviceID,
		Name:        name,
	})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	d.SetId(utils.MarshalID(serviceID, fe.Name))

	if err = d.Set("loadbalancer", serviceID); err != nil {
		return diag.FromErr(err)
	}

	if diags = setFrontendResourceData(d, fe); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceFrontendUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &name); err != nil {
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
			Properties:     frontendPropertiesFromResourceData(d),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(d.Get("loadbalancer").(string), fe.Name))

	if diags = setFrontendResourceData(d, fe); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "frontend updated", map[string]interface{}{"name": fe.Name, "service_uuid": serviceID})
	return diags
}

func resourceFrontendDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "deleting frontend", map[string]interface{}{"name": name, "service_uuid": serviceID})
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

	if fe.Properties != nil {
		props := []map[string]interface{}{{
			"timeout_client":         fe.Properties.TimeoutClient,
			"inbound_proxy_protocol": fe.Properties.InboundProxyProtocol,
		}}
		if err := d.Set("properties", props); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func frontendPropertiesFromResourceData(d *schema.ResourceData) *upcloud.LoadBalancerFrontendProperties {
	if props, ok := d.GetOk("properties.0"); !ok || props == nil {
		return nil
	}
	return &upcloud.LoadBalancerFrontendProperties{
		TimeoutClient:        d.Get("properties.0.timeout_client").(int),
		InboundProxyProtocol: d.Get("properties.0.inbound_proxy_protocol").(bool),
	}
}

func schemaFrontendProperties() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"timeout_client": {
			Description:      "Client request timeout in seconds.",
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          10,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 86400)),
		},
		"inbound_proxy_protocol": {
			Description: "Enable or disable inbound proxy protocol support.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
	}
}

func loadBalancerFrontendNetworksFromResourceData(d *schema.ResourceData) []upcloud.LoadBalancerFrontendNetwork {
	req := make([]upcloud.LoadBalancerFrontendNetwork, 0)
	if nets, ok := d.GetOk("networks"); ok {
		for _, n := range nets.([]interface{}) {
			n := n.(map[string]interface{})
			req = append(req, upcloud.LoadBalancerFrontendNetwork{
				Name: n["name"].(string),
			})
		}
	}
	return req
}
