package loadbalancer

import (
	"context"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceManualCertificateBundle() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents manual certificate bundle",
		CreateContext: resourceManualCertificateBundleCreate,
		ReadContext:   resourceManualCertificateBundleRead,
		UpdateContext: resourceManualCertificateBundleUpdate,
		DeleteContext: resourceCertificateBundleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description:      "The name of the bundle must be unique within customer account.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateNameDiagFunc,
			},
			"certificate": {
				Description: "Certificate within base64 string must be in PEM format.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"intermediates": {
				Description: "Intermediate certificates within base64 string must be in PEM format.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"private_key": {
				Description: "Private key within base64 string must be in PEM format.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"operational_state": {
				Description: "The service operational state indicates the service's current operational, effective state. Managed by the system.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"not_after": {
				Description: "The time after which a certificate is no longer valid.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"not_before": {
				Description: "The time on which a certificate becomes valid.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceManualCertificateBundleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	b, err := svc.CreateLoadBalancerCertificateBundle(ctx, &request.CreateLoadBalancerCertificateBundleRequest{
		Type:          upcloud.LoadBalancerCertificateBundleTypeManual,
		Name:          d.Get("name").(string),
		Certificate:   d.Get("certificate").(string),
		Intermediates: d.Get("intermediates").(string),
		PrivateKey:    d.Get("private_key").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(b.UUID)

	if diags = setManualCertificateBundleResourceData(d, b); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "certificate bundle created", map[string]interface{}{"name": b.Name, "uuid": b.UUID})
	return diags
}

func resourceManualCertificateBundleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	b, err := svc.GetLoadBalancerCertificateBundle(ctx, &request.GetLoadBalancerCertificateBundleRequest{
		UUID: d.Id(),
	})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	if diags = setManualCertificateBundleResourceData(d, b); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceManualCertificateBundleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	b, err := svc.ModifyLoadBalancerCertificateBundle(ctx, &request.ModifyLoadBalancerCertificateBundleRequest{
		UUID:          d.Id(),
		Name:          d.Get("name").(string),
		Certificate:   d.Get("certificate").(string),
		Intermediates: upcloud.StringPtr(d.Get("intermediates").(string)),
		PrivateKey:    d.Get("private_key").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if diags = setManualCertificateBundleResourceData(d, b); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "certificate bundle updated", map[string]interface{}{"name": b.Name, "uuid": b.UUID})
	return diags
}

func resourceCertificateBundleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	tflog.Info(ctx, "deleting certificate bundle", map[string]interface{}{"name": d.Get("name").(string), "uuid": d.Id()})
	return diag.FromErr(svc.DeleteLoadBalancerCertificateBundle(ctx, &request.DeleteLoadBalancerCertificateBundleRequest{UUID: d.Id()}))
}

func setManualCertificateBundleResourceData(d *schema.ResourceData, b *upcloud.LoadBalancerCertificateBundle) (diags diag.Diagnostics) {
	if err := d.Set("name", b.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("certificate", b.Certificate); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("intermediates", b.Intermediates); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("operational_state", b.OperationalState); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("not_after", b.NotAfter.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("not_before", b.NotAfter.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
