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

func ResourceFrontendTLSConfig() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents frontend TLS config",
		CreateContext: resourceFrontendTLSConfigCreate,
		ReadContext:   resourceFrontendTLSConfigRead,
		UpdateContext: resourceFrontendTLSConfigUpdate,
		DeleteContext: resourceFrontendTLSConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"frontend": {
				Description: "ID of the load balancer frontend to which the TLS config is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description:      "The name of the TLS config must be unique within service frontend.",
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

func resourceFrontendTLSConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, feName string
	if err := utils.UnmarshalID(d.Get("frontend").(string), &serviceID, &feName); err != nil {
		return diag.FromErr(err)
	}
	t, err := svc.CreateLoadBalancerFrontendTLSConfig(ctx, &request.CreateLoadBalancerFrontendTLSConfigRequest{
		ServiceUUID:  serviceID,
		FrontendName: feName,
		Config: request.LoadBalancerFrontendTLSConfig{
			Name:                  d.Get("name").(string),
			CertificateBundleUUID: d.Get("certificate_bundle").(string),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(serviceID, feName, t.Name))

	if diags = setFrontendTLSConfigResourceData(d, t); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "frontend TLS config created", map[string]interface{}{"name": t.Name, "service_uuid": serviceID, "fe_name": feName})
	return diags
}

func resourceFrontendTLSConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, feName, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &feName, &name); err != nil {
		return diag.FromErr(err)
	}
	t, err := svc.GetLoadBalancerFrontendTLSConfig(ctx, &request.GetLoadBalancerFrontendTLSConfigRequest{
		ServiceUUID:  serviceID,
		FrontendName: feName,
		Name:         name,
	})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	if err = d.Set("frontend", utils.MarshalID(serviceID, feName)); err != nil {
		return diag.FromErr(err)
	}

	if diags = setFrontendTLSConfigResourceData(d, t); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceFrontendTLSConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, feName, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &feName, &name); err != nil {
		return diag.FromErr(err)
	}
	t, err := svc.ModifyLoadBalancerFrontendTLSConfig(ctx, &request.ModifyLoadBalancerFrontendTLSConfigRequest{
		ServiceUUID:  serviceID,
		FrontendName: feName,
		Name:         name,
		Config: request.LoadBalancerFrontendTLSConfig{
			Name:                  d.Get("name").(string),
			CertificateBundleUUID: d.Get("certificate_bundle").(string),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(serviceID, feName, t.Name))

	if diags = setFrontendTLSConfigResourceData(d, t); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "frontend TLS config updated", map[string]interface{}{"name": t.Name, "service_uuid": serviceID, "fe_name": feName})
	return diags
}

func resourceFrontendTLSConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, feName, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &feName, &name); err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "deleting frontend TLS config", map[string]interface{}{"name": name, "service_uuid": serviceID, "fe_name": feName})
	return diag.FromErr(svc.DeleteLoadBalancerFrontendTLSConfig(ctx, &request.DeleteLoadBalancerFrontendTLSConfigRequest{
		ServiceUUID:  serviceID,
		FrontendName: feName,
		Name:         name,
	}))
}

func setFrontendTLSConfigResourceData(d *schema.ResourceData, t *upcloud.LoadBalancerFrontendTLSConfig) (diags diag.Diagnostics) {
	if err := d.Set("name", t.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("certificate_bundle", t.CertificateBundleUUID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
