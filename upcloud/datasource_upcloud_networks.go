package upcloud

import (
	"context"
	"regexp"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetworks() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworksRead,
		Schema: map[string]*schema.Schema{
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `If specified, this data source will return only networks from this zone`,
			},
			"filter_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `If specified, results will be filtered to match name using a regular expression`,
			},
			"networks": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_network": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "A list of IP subnets within the network",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"address": {
										Type:        schema.TypeString,
										Description: "The CIDR range of the subnet",
										Computed:    true,
									},
									"dhcp": {
										Type:        schema.TypeBool,
										Description: "Is DHCP enabled?",
										Computed:    true,
									},
									"dhcp_default_route": {
										Type:        schema.TypeBool,
										Description: "Is the gateway the DHCP default route?",
										Computed:    true,
									},
									"dhcp_dns": {
										Type:        schema.TypeList,
										Description: "The DNS servers given by DHCP",
										Computed:    true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"family": {
										Type:        schema.TypeString,
										Description: "IP address family",
										Computed:    true,
									},
									"gateway": {
										Type:        schema.TypeString,
										Description: "Gateway address given by DHCP",
										Computed:    true,
									},
								},
							},
						},
						"name": {
							Type:        schema.TypeString,
							Description: "A valid name for the network",
							Computed:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "The network type",
							Computed:    true,
						},
						"id": {
							Type:        schema.TypeString,
							Description: "The UUID of the network",
							Computed:    true,
						},
						"zone": {
							Type:        schema.TypeString,
							Description: "The zone the network is in",
							Computed:    true,
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
				},
			},
		},
	}
}

func dataSourceNetworksRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	// Get the zone from the configuration
	var zone string
	if z := d.Get("zone"); z != nil {
		zone = z.(string)
	}

	var filterName string
	if fn := d.Get("filter_name"); fn != nil {
		filterName = fn.(string)
	}

	// Fetch the networks from the API. If "zone" is specified in configuration
	// use the `GetNetworksInZone` function.
	var err error
	var fetchedNetworks *upcloud.Networks
	if zone != "" {
		fetchedNetworks, err = client.GetNetworksInZone(&request.GetNetworksInZoneRequest{
			Zone: zone,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		fetchedNetworks, err = client.GetNetworks()
		if err != nil {
			return diag.FromErr(err)
		}
	}

	filteredNetworks := fetchedNetworks.Networks
	if filterName != "" {
		filteredNetworks, err = FilterNetworks(fetchedNetworks.Networks, func(n upcloud.Network) (bool, error) {
			return regexp.MatchString(filterName, n.Name)
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Map the received data to the Terraform resource.
	var networks []map[string]interface{}
	for _, fn := range filteredNetworks {
		n := map[string]interface{}{
			"name": fn.Name,
			"type": fn.Type,
			"id":   fn.UUID,
			"zone": fn.Zone,
		}

		var ipns []map[string]interface{}
		for _, fipn := range fn.IPNetworks {
			ipn := map[string]interface{}{
				"address":            fipn.Address,
				"dhcp":               fipn.DHCP.Bool(),
				"dhcp_default_route": fipn.DHCPDefaultRoute.Bool(),
				"dhcp_dns":           fipn.DHCPDns,
				"family":             fipn.Family,
				"gateway":            fipn.Gateway,
			}

			ipns = append(ipns, ipn)
		}

		n["ip_network"] = ipns

		networks = append(networks, n)
	}

	err = d.Set("networks", networks)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(time.Now().UTC().Format(time.RFC3339Nano))

	return nil
}
