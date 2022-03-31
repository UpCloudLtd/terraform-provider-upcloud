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
	if err := unmarshalID(d.Get("frontend").(string), &serviceID, &feName); err != nil {
		return diag.FromErr(err)
	}
	t, err := svc.CreateLoadBalancerFrontendTLSConfig(&request.CreateLoadBalancerFrontendTLSConfigRequest{
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

	d.SetId(marshalID(serviceID, feName, t.Name))

	if diags = setFrontendTLSConfigResourceData(d, t); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] frontend TLS config '%s' created", t.Name)
	return diags
}

func resourceFrontendTLSConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, feName, name string
	if err := unmarshalID(d.Id(), &serviceID, &feName, &name); err != nil {
		return diag.FromErr(err)
	}
	t, err := svc.GetLoadBalancerFrontendTLSConfig(&request.GetLoadBalancerFrontendTLSConfigRequest{
		ServiceUUID:  serviceID,
		FrontendName: feName,
		Name:         name,
	})

	if err != nil {
		return handleResourceError(d.Get("name").(string), d, err)
	}

	if diags = setFrontendTLSConfigResourceData(d, t); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceFrontendTLSConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, feName, name string
	if err := unmarshalID(d.Id(), &serviceID, &feName, &name); err != nil {
		return diag.FromErr(err)
	}
	t, err := svc.ModifyLoadBalancerFrontendTLSConfig(&request.ModifyLoadBalancerFrontendTLSConfigRequest{
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

	d.SetId(marshalID(serviceID, feName, t.Name))

	if diags = setFrontendTLSConfigResourceData(d, t); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] frontend TLS config '%s' updated", t.Name)
	return diags
}

func resourceFrontendTLSConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, feName, name string
	if err := unmarshalID(d.Id(), &serviceID, &feName, &name); err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] deleting frontend TLS config '%s'", d.Id())
	return diag.FromErr(svc.DeleteLoadBalancerFrontendTLSConfig(&request.DeleteLoadBalancerFrontendTLSConfigRequest{
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
