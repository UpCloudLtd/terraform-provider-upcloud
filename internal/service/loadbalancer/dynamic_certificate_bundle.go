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

func ResourceDynamicCertificateBundle() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents dynamic certificate bundle",
		CreateContext: resourceDynamicCertificateBundleCreate,
		ReadContext:   resourceDynamicCertificateBundleRead,
		UpdateContext: resourceDynamicCertificateBundleUpdate,
		DeleteContext: resourceDynamicCertificateBundleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the bundle must be unique within customer account.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hostnames": {
				Description: "Certificate hostnames.",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    100,
				MinItems:    1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"key_type": {
				Description: "Private key type (`rsa` / `ecdsa`).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice([]string{"rsa", "ecdsa"}, false)),
			},
		},
	}
}

func resourceDynamicCertificateBundleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	hostnames := make([]string, 0)
	for _, h := range d.Get("hostnames").([]interface{}) {
		hostnames = append(hostnames, h.(string))
	}
	b, err := svc.CreateLoadBalancerCertificateBundle(&request.CreateLoadBalancerCertificateBundleRequest{
		Type:      upcloud.LoadBalancerCertificateBundleTypeDynamic,
		Name:      d.Get("name").(string),
		KeyType:   d.Get("key_type").(string),
		Hostnames: hostnames,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(b.UUID)

	if diags = setDynamicCertificateBundleResourceData(d, b); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] certificate bundle '%s' created", b.Name)
	return diags
}

func resourceDynamicCertificateBundleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	b, err := svc.GetLoadBalancerCertificateBundle(&request.GetLoadBalancerCertificateBundleRequest{
		UUID: d.Id(),
	})

	if err != nil {
		return handleResourceError(d.Get("name").(string), d, err)
	}

	if diags = setDynamicCertificateBundleResourceData(d, b); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceDynamicCertificateBundleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	hostnames := make([]string, 0)
	for _, h := range d.Get("hostnames").([]interface{}) {
		hostnames = append(hostnames, h.(string))
	}
	b, err := svc.ModifyLoadBalancerCertificateBundle(&request.ModifyLoadBalancerCertificateBundleRequest{
		UUID:      d.Id(),
		Name:      d.Get("name").(string),
		Hostnames: hostnames,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if diags = setDynamicCertificateBundleResourceData(d, b); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] certificate bundle '%s' updated", b.Name)
	return diags
}

func resourceDynamicCertificateBundleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	log.Printf("[INFO] deleting certificate bundle '%s' (%s)", d.Get("name").(string), d.Id())
	return diag.FromErr(svc.DeleteLoadBalancerCertificateBundle(&request.DeleteLoadBalancerCertificateBundleRequest{UUID: d.Id()}))
}

func setDynamicCertificateBundleResourceData(d *schema.ResourceData, b *upcloud.LoadBalancerCertificateBundle) (diags diag.Diagnostics) {
	if err := d.Set("name", b.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("key_type", b.KeyType); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("hostnames", b.Hostnames); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
