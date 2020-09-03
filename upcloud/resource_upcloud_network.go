package upcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/go-cty/cty"
)

func resourceUpCloudNetwork() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceUpCloudNetworkRead,
		CreateContext: resourceUpCloudNetworkCreate,
		UpdateContext: resourceUpCloudNetworkUpdate,
		DeleteContext: resourceUpCloudNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"ip_network": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				MinItems:    1,
				Description: "A list of IP subnets within the network",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:         schema.TypeString,
							Description:  "The CIDR range of the subnet",
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsCIDR,
						},
						"dhcp": {
							Type:        schema.TypeBool,
							Description: "Is DHCP enabled?",
							Required:    true,
						},
						"dhcp_default_route": {
							Type:        schema.TypeBool,
							Description: "Is the gateway the DHCP default route?",
							Computed:    true,
							Optional:    true,
						},
						"dhcp_dns": {
							Type:        schema.TypeSet,
							Description: "The DNS servers given by DHCP",
							Computed:    true,
							Optional:    true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.Any(validation.IsIPv4Address, validation.IsIPv6Address),
							},
						},
						"family": {
							Type:        schema.TypeString,
							Description: "IP address family",
							Required:    true,
							ValidateDiagFunc: func(v interface{}, _ cty.Path) diag.Diagnostics {
								switch v.(string) {
								case upcloud.IPAddressFamilyIPv4, upcloud.IPAddressFamilyIPv6:
									return nil
								default:
									return diag.Diagnostics{diag.Diagnostic{
										Severity: diag.Error,
										Summary:  "'family' has incorrect value",
										Detail: fmt.Sprintf("'family' should have value of %s or %s",
											upcloud.IPAddressFamilyIPv4,
											upcloud.IPAddressFamilyIPv6),
									}}
								}
							},
						},
						"gateway": {
							Type:        schema.TypeString,
							Description: "Gateway address given by DHCP",
							Computed:    true,
							Optional:    true,
						},
					},
				},
			},
			"name": {
				Type:        schema.TypeString,
				Description: "A valid name for the network",
				Required:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The network type",
				Computed:    true,
			},
			"zone": {
				Type:        schema.TypeString,
				Description: "The zone the network is in",
				Required:    true,
				ForceNew:    true,
			},
			"router": {
				Type:        schema.TypeString,
				Description: "The UUID of a router",
				Optional:    true,
			},
			"servers": {
				Type:        schema.TypeSet,
				Description: "A list of attached servers",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "The UUID of the server",
							Computed:    true,
						},
						"title": {
							Type:        schema.TypeString,
							Description: "The short description of the server",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func resourceUpCloudNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	req := request.CreateNetworkRequest{}
	if v := d.Get("name"); v != nil {
		req.Name = v.(string)
	}

	if v := d.Get("zone"); v != nil {
		req.Zone = v.(string)
	}

	if v := d.Get("router"); v != nil {
		req.Router = v.(string)
	}

	if v, ok := d.GetOk("ip_network"); ok {
		ipn := v.([]interface{})[0]
		ipnConf := ipn.(map[string]interface{})

		uipn := upcloud.IPNetwork{
			Address:          ipnConf["address"].(string),
			DHCP:             upcloud.FromBool(ipnConf["dhcp"].(bool)),
			DHCPDefaultRoute: upcloud.FromBool(ipnConf["dhcp_default_route"].(bool)),
			Family:           ipnConf["family"].(string),
			Gateway:          ipnConf["gateway"].(string),
		}

		for _, dns := range ipnConf["dhcp_dns"].(*schema.Set).List() {
			uipn.DHCPDns = append(uipn.DHCPDns, dns.(string))
		}

		req.IPNetworks = append(req.IPNetworks, uipn)
	}

	network, err := client.CreateNetwork(&req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(network.UUID)

	return resourceUpCloudNetworkRead(ctx, d, meta)
}

func resourceUpCloudNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	req := request.GetNetworkDetailsRequest{
		UUID: d.Id(),
	}

	network, err := client.GetNetworkDetails(&req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", network.Name)
	d.Set("type", network.Type)
	d.Set("zone", network.Zone)
	if network.Router != "" {
		d.Set("router", network.Router)
	}

	if len(network.IPNetworks) > 1 {
		return diag.Errorf("too many ip_networks: %d", len(network.IPNetworks))
	}

	if len(network.IPNetworks) == 1 {
		ipn := map[string]interface{}{
			"address":            network.IPNetworks[0].Address,
			"dhcp":               network.IPNetworks[0].DHCP.Bool(),
			"dhcp_default_route": network.IPNetworks[0].DHCPDefaultRoute.Bool(),
			"dhcp_dns":           network.IPNetworks[0].DHCPDns,
			"family":             network.IPNetworks[0].Family,
			"gateway":            network.IPNetworks[0].Gateway,
		}

		d.Set("ip_network", []map[string]interface{}{
			ipn,
		})
	}

	d.SetId(network.UUID)

	return nil
}

func resourceUpCloudNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	req := request.ModifyNetworkRequest{
		UUID: d.Id(),
	}

	if d.HasChange("name") {
		_, v := d.GetChange("name")
		req.Name = v.(string)
	}

	if d.HasChange("router") {
		_, v := d.GetChange("router")
		req.Router = v.(string)
	}

	if d.HasChange("ip_network") {
		v := d.Get("ip_network")

		ipn := v.([]interface{})[0]
		ipnConf := ipn.(map[string]interface{})

		uipn := upcloud.IPNetwork{
			Address:          ipnConf["address"].(string),
			DHCP:             upcloud.FromBool(ipnConf["dhcp"].(bool)),
			DHCPDefaultRoute: upcloud.FromBool(ipnConf["dhcp_default_route"].(bool)),
			Family:           ipnConf["family"].(string),
			Gateway:          ipnConf["gateway"].(string),
		}

		for _, dns := range ipnConf["dhcp_dns"].(*schema.Set).List() {
			uipn.DHCPDns = append(uipn.DHCPDns, dns.(string))
		}

		req.IPNetworks = []upcloud.IPNetwork{uipn}
	}

	network, err := client.ModifyNetwork(&req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(network.UUID)

	return nil
}

func resourceUpCloudNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	req := request.DeleteNetworkRequest{
		UUID: d.Id(),
	}
	err := client.DeleteNetwork(&req)

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
