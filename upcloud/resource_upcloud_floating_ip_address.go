package upcloud

import (
	"context"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUpCloudFloatingIPAddress() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUpCloudFloatingIPAddressCreate,
		ReadContext:   resourceUpCloudFloatingIPAddressRead,
		UpdateContext: resourceUpCloudFloatingIPAddressUpdate,
		DeleteContext: resourceUpCloudFloatingIPAddressDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"ip_address": {
				Description: "An UpCloud assigned IP Address",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"access": {
				Description:  "Is address for utility or public network",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "public",
				ValidateFunc: validation.StringInSlice([]string{"utility", "public"}, false),
			},
			"family": {
				Description:  "The address family of new IP address",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "IPv4",
				ValidateFunc: validation.StringInSlice([]string{"IPv4", "IPv6"}, false),
			},
			"mac_address": {
				Description:  "MAC address of server interface to assign address to",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsMACAddress,
			},
			"zone": {
				Description: "Zone of address, required when assigning a detached floating IP address",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceUpCloudFloatingIPAddressCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	assignIPAddressRequest := &request.AssignIPAddressRequest{
		Floating: upcloud.True,
	}

	if access, ok := d.GetOk("access"); ok {
		assignIPAddressRequest.Access = access.(string)
	}
	if family, ok := d.GetOk("Family"); ok {
		assignIPAddressRequest.Family = family.(string)
	}
	if mac, ok := d.GetOk("mac_address"); ok {
		assignIPAddressRequest.MAC = mac.(string)
	}
	if zone, ok := d.GetOk("zone"); ok {
		assignIPAddressRequest.Zone = zone.(string)
	}

	ipAddress, err := client.AssignIPAddress(assignIPAddressRequest)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ip_address", ipAddress.Address); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ipAddress.Address)

	return resourceUpCloudFloatingIPAddressRead(ctx, d, meta)
}

func resourceUpCloudFloatingIPAddressRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	getIPAddressDetailsRequest := &request.GetIPAddressDetailsRequest{
		Address: d.Id(),
	}

	ipAddress, err := client.GetIPAddressDetails(getIPAddressDetailsRequest)

	if err != nil {
		diag.FromErr(err)
	}

	if err := d.Set("ip_address", ipAddress.Address); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("access", ipAddress.Access); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("family", ipAddress.Family); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("mac_address", ipAddress.MAC); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("zone", ipAddress.Zone); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceUpCloudFloatingIPAddressUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	modifyIPAddressRequest := &request.ModifyIPAddressRequest{
		IPAddress: d.Id(),
	}

	if d.HasChange("mac_address") {
		_, newMAC := d.GetChange("mac_address")
		modifyIPAddressRequest.MAC = newMAC.(string)
	}

	_, err := client.ModifyIPAddress(modifyIPAddressRequest)
	if err != nil {
		diag.FromErr(err)
	}

	return resourceUpCloudFloatingIPAddressRead(ctx, d, meta)
}

func resourceUpCloudFloatingIPAddressDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	if _, ok := d.GetOk("mac_address"); ok {

		modifyIPAddressRequest := &request.ModifyIPAddressRequest{
			IPAddress: d.Id(),
			MAC:       "",
		}

		_, err := client.ModifyIPAddress(modifyIPAddressRequest)
		if err != nil {
			diag.FromErr(err)
		}
	}

	releaseIPAddressRequest := &request.ReleaseIPAddressRequest{
		IPAddress: d.Id(),
	}

	err := client.ReleaseIPAddress(releaseIPAddressRequest)

	if err != nil {
		diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
