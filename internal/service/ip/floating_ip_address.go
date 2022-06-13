package ip

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const ipAddressNotFoundErrorCode string = "IP_ADDRESS_NOT_FOUND"

func ResourceFloatingIPAddress() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents a UpCloud floating IP address resource.",
		CreateContext: resourceFloatingIPAddressCreate,
		ReadContext:   resourceFloatingIPAddressRead,
		UpdateContext: resourceFloatingIPAddressUpdate,
		DeleteContext: resourceFloatingIPAddressDelete,
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

func resourceFloatingIPAddressCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.ServiceContext)

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

	ipAddress, err := client.AssignIPAddress(ctx, assignIPAddressRequest)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ip_address", ipAddress.Address); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ipAddress.Address)

	return resourceFloatingIPAddressRead(ctx, d, meta)
}

func resourceFloatingIPAddressRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.ServiceContext)

	var diags diag.Diagnostics

	getIPAddressDetailsRequest := &request.GetIPAddressDetailsRequest{
		Address: d.Id(),
	}

	ipAddress, err := client.GetIPAddressDetails(ctx, getIPAddressDetailsRequest)

	if err != nil {
		if svcErr, ok := err.(*upcloud.Error); ok && svcErr.ErrorCode == ipAddressNotFoundErrorCode {
			name := "ip address" // set default name because ip_address is optional field
			if ip, ok := d.GetOk("ip_address"); ok {
				name = ip.(string)
			}
			diags = append(diags, utils.DiagBindingRemovedWarningFromUpcloudErr(svcErr, name))
			d.SetId("")
			return diags
		}
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

func resourceFloatingIPAddressUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.ServiceContext)

	modifyIPAddressRequest := &request.ModifyIPAddressRequest{
		IPAddress: d.Id(),
	}

	if d.HasChange("mac_address") {
		_, newMAC := d.GetChange("mac_address")
		modifyIPAddressRequest.MAC = newMAC.(string)
	}

	_, err := client.ModifyIPAddress(ctx, modifyIPAddressRequest)
	if err != nil {
		diag.FromErr(err)
	}

	return resourceFloatingIPAddressRead(ctx, d, meta)
}

func resourceFloatingIPAddressDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.ServiceContext)

	var diags diag.Diagnostics

	if _, ok := d.GetOk("mac_address"); ok {
		modifyIPAddressRequest := &request.ModifyIPAddressRequest{
			IPAddress: d.Id(),
			MAC:       "",
		}

		_, err := client.ModifyIPAddress(ctx, modifyIPAddressRequest)
		if err != nil {
			diag.FromErr(err)
		}
	}

	releaseIPAddressRequest := &request.ReleaseIPAddressRequest{
		IPAddress: d.Id(),
	}

	err := client.ReleaseIPAddress(ctx, releaseIPAddressRequest)

	if err != nil {
		diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
