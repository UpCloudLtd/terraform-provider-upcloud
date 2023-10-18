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
			"properties": {
				Description: "Backend properties. Properties can set back to defaults by defining empty `properties {}` block.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: schemaBackendProperties(),
				},
			},
		},
	}
}

func resourceBackendCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	serviceID := d.Get("loadbalancer").(string)

	be, err := svc.CreateLoadBalancerBackend(ctx, &request.CreateLoadBalancerBackendRequest{
		ServiceUUID: serviceID,
		Backend: request.LoadBalancerBackend{
			Name:       d.Get("name").(string),
			Resolver:   d.Get("resolver_name").(string),
			Members:    []request.LoadBalancerBackendMember{},
			Properties: backendPropertiesFromResourceData(d),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(serviceID, be.Name))

	if diags = setBackendResourceData(d, be); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "backend created", map[string]interface{}{"name": be.Name, "service_uuid": serviceID})
	return diags
}

func resourceBackendRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	be, err := svc.GetLoadBalancerBackend(ctx, &request.GetLoadBalancerBackendRequest{
		ServiceUUID: serviceID,
		Name:        name,
	})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	d.SetId(utils.MarshalID(serviceID, be.Name))

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
	if err := utils.UnmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}

	be, err := svc.ModifyLoadBalancerBackend(ctx, &request.ModifyLoadBalancerBackendRequest{
		ServiceUUID: serviceID,
		Name:        name,
		Backend: request.ModifyLoadBalancerBackend{
			Name:       d.Get("name").(string),
			Resolver:   upcloud.StringPtr(d.Get("resolver_name").(string)),
			Properties: backendPropertiesFromResourceData(d),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(d.Get("loadbalancer").(string), be.Name))

	if diags = setBackendResourceData(d, be); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "backend updated", map[string]interface{}{"name": be.Name, "service_uuid": serviceID})
	return diags
}

func resourceBackendDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "deleting backend", map[string]interface{}{"name": name, "service_uuid": serviceID})
	return diag.FromErr(svc.DeleteLoadBalancerBackend(ctx, &request.DeleteLoadBalancerBackendRequest{
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

	if be.Properties != nil {
		props := []map[string]interface{}{{
			"timeout_server":               be.Properties.TimeoutServer,
			"timeout_tunnel":               be.Properties.TimeoutTunnel,
			"health_check_type":            be.Properties.HealthCheckType,
			"health_check_interval":        be.Properties.HealthCheckInterval,
			"health_check_fall":            be.Properties.HealthCheckFall,
			"health_check_rise":            be.Properties.HealthCheckRise,
			"health_check_url":             be.Properties.HealthCheckURL,
			"health_check_tls_verify":      be.Properties.HealthCheckTLSVerify,
			"health_check_expected_status": be.Properties.HealthCheckExpectedStatus,
			"sticky_session_cookie_name":   be.Properties.StickySessionCookieName,
			"outbound_proxy_protocol":      be.Properties.OutboundProxyProtocol,
		}}
		if err := d.Set("properties", props); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func backendPropertiesFromResourceData(d *schema.ResourceData) *upcloud.LoadBalancerBackendProperties {
	if props, ok := d.GetOk("properties.0"); !ok || props == nil {
		return nil
	}
	return &upcloud.LoadBalancerBackendProperties{
		TimeoutServer:             d.Get("properties.0.timeout_server").(int),
		TimeoutTunnel:             d.Get("properties.0.timeout_tunnel").(int),
		HealthCheckType:           upcloud.LoadBalancerHealthCheckType(d.Get("properties.0.health_check_type").(string)),
		HealthCheckInterval:       d.Get("properties.0.health_check_interval").(int),
		HealthCheckFall:           d.Get("properties.0.health_check_fall").(int),
		HealthCheckRise:           d.Get("properties.0.health_check_rise").(int),
		HealthCheckURL:            d.Get("properties.0.health_check_url").(string),
		HealthCheckTLSVerify:      upcloud.BoolPtr(d.Get("properties.0.health_check_tls_verify").(bool)),
		HealthCheckExpectedStatus: d.Get("properties.0.health_check_expected_status").(int),
		StickySessionCookieName:   d.Get("properties.0.sticky_session_cookie_name").(string),
		OutboundProxyProtocol:     upcloud.LoadBalancerProxyProtocolVersion(d.Get("properties.0.outbound_proxy_protocol").(string)),
	}
}

func schemaBackendProperties() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"timeout_server": {
			Description:      "Backend server timeout in seconds.",
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          10,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 86400)),
		},
		"timeout_tunnel": {
			Description:      "Maximum inactivity time on the client and server side for tunnels in seconds.",
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          3600,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 3024000)),
		},
		"health_check_type": {
			Description: "Health check type.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "tcp",
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice([]string{
					string(upcloud.LoadBalancerHealthCheckTypeTCP),
					string(upcloud.LoadBalancerHealthCheckTypeHTTP),
				}, false),
			),
		},
		"health_check_interval": {
			Description:      "Interval between health checks.",
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          10,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 86400)),
		},
		"health_check_fall": {
			Description:      "Sets how many failed health checks are allowed until the backend member is taken off from the rotation.",
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          3,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 100)),
		},
		"health_check_rise": {
			Description:      "Sets how many passing checks there must be before returning the backend member to the rotation.",
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          3,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 100)),
		},
		"health_check_url": {
			Description:      "Target path for health check HTTP GET requests. Ignored for tcp type.",
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "/",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 255)),
		},
		"health_check_tls_verify": {
			Description: "Enables certificate verification with the system CA certificate bundle. Works with https scheme in health_check_url, otherwise ignored.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"health_check_expected_status": {
			Description:      "Expected HTTP status code returned by the customer application to mark server as healthy. Ignored for tcp type.",
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          200,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(100, 599)),
		},
		"sticky_session_cookie_name": {
			Description:      "Sets sticky session cookie name. Empty string disables sticky session.",
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 64)),
		},
		"outbound_proxy_protocol": {
			Description: "Enable outbound proxy protocol by setting the desired version. Empty string disables proxy protocol.",
			Type:        schema.TypeString,
			Optional:    true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice([]string{
					"",
					string(upcloud.LoadBalancerProxyProtocolVersion1),
					string(upcloud.LoadBalancerProxyProtocolVersion2),
				}, false),
			),
		},
	}
}
