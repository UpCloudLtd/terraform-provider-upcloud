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
)

func ResourceBackendTLSConfig() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents backend TLS config",
		CreateContext: resourceBackendTLSConfigCreate,
		ReadContext:   resourceBackendTLSConfigRead,
		UpdateContext: resourceBackendTLSConfigUpdate,
		DeleteContext: resourceBackendTLSConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"backend": {
				Description: "ID of the load balancer backend to which the TLS config is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description:      "The name of the TLS config must be unique within service backend.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateNameDiagFunc,
			},
			"certificate_bundle": {
				Description: "Reference to certificate bundle ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceBackendTLSConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, beName string
	if err := utils.UnmarshalID(d.Get("backend").(string), &serviceID, &beName); err != nil {
		return diag.FromErr(err)
	}
	t, err := svc.CreateLoadBalancerBackendTLSConfig(ctx, &request.CreateLoadBalancerBackendTLSConfigRequest{
		ServiceUUID: serviceID,
		BackendName: beName,
		Config: request.LoadBalancerBackendTLSConfig{
			Name:                  d.Get("name").(string),
			CertificateBundleUUID: d.Get("certificate_bundle").(string),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(serviceID, beName, t.Name))

	if diags = setBackendTLSConfigResourceData(d, t); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "backend TLS config created", map[string]interface{}{"name": t.Name, "service_uuid": serviceID, "be_name": beName})
	return diags
}

func resourceBackendTLSConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, beName, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &beName, &name); err != nil {
		return diag.FromErr(err)
	}
	t, err := svc.GetLoadBalancerBackendTLSConfig(ctx, &request.GetLoadBalancerBackendTLSConfigRequest{
		ServiceUUID: serviceID,
		BackendName: beName,
		Name:        name,
	})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	if err = d.Set("backend", utils.MarshalID(serviceID, beName)); err != nil {
		return diag.FromErr(err)
	}

	if diags = setBackendTLSConfigResourceData(d, t); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceBackendTLSConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, beName, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &beName, &name); err != nil {
		return diag.FromErr(err)
	}
	t, err := svc.ModifyLoadBalancerBackendTLSConfig(ctx, &request.ModifyLoadBalancerBackendTLSConfigRequest{
		ServiceUUID: serviceID,
		BackendName: beName,
		Name:        name,
		Config: request.LoadBalancerBackendTLSConfig{
			Name:                  d.Get("name").(string),
			CertificateBundleUUID: d.Get("certificate_bundle").(string),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(serviceID, beName, t.Name))

	if diags = setBackendTLSConfigResourceData(d, t); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "backend TLS config updated", map[string]interface{}{"name": t.Name, "service_uuid": serviceID, "be_name": beName})
	return diags
}

func resourceBackendTLSConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, beName, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &beName, &name); err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "deleting backend TLS config", map[string]interface{}{"name": name, "service_uuid": serviceID, "be_name": beName})
	return diag.FromErr(svc.DeleteLoadBalancerBackendTLSConfig(ctx, &request.DeleteLoadBalancerBackendTLSConfigRequest{
		ServiceUUID: serviceID,
		BackendName: beName,
		Name:        name,
	}))
}

func setBackendTLSConfigResourceData(d *schema.ResourceData, t *upcloud.LoadBalancerBackendTLSConfig) (diags diag.Diagnostics) {
	if err := d.Set("name", t.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("certificate_bundle", t.CertificateBundleUUID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
