package loadbalancer

import (
	"context"
	"log"
	"time"

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

func resourceDynamicCertificateBundleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)
	hostnames := make([]string, 0)
	for _, h := range d.Get("hostnames").([]interface{}) {
		hostnames = append(hostnames, h.(string))
	}
	b, err := svc.CreateLoadBalancerCertificateBundle(ctx, &request.CreateLoadBalancerCertificateBundleRequest{
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
	svc := meta.(*service.ServiceContext)
	b, err := svc.GetLoadBalancerCertificateBundle(ctx, &request.GetLoadBalancerCertificateBundleRequest{
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
	svc := meta.(*service.ServiceContext)
	hostnames := make([]string, 0)
	for _, h := range d.Get("hostnames").([]interface{}) {
		hostnames = append(hostnames, h.(string))
	}
	b, err := svc.ModifyLoadBalancerCertificateBundle(ctx, &request.ModifyLoadBalancerCertificateBundleRequest{
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
