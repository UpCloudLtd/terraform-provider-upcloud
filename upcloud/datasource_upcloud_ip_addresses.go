package upcloud

import (
	"context"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUpCloudIPAddresses() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUpCloudIPAddressesRead,
		Schema: map[string]*schema.Schema{
			"addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access": {
							Type:        schema.TypeString,
							Description: "Is address for utility or public network",
							Computed:    true,
						},
						"address": {
							Type:        schema.TypeString,
							Description: "An UpCloud assigned IP Address",
							Computed:    true,
						},
						"family": {
							Type:        schema.TypeString,
							Description: "IP address family",
							Computed:    true,
						},
						"part_of_plan": {
							Type:        schema.TypeBool,
							Description: "Is the address a part of a plan",
							Computed:    true,
						},
						"ptr_record": {
							Type:        schema.TypeString,
							Description: "A reverse DNS record entry",
							Computed:    true,
						},
						"server": {
							Type:        schema.TypeString,
							Description: "The unique identifier for a server",
							Computed:    true,
						},
						"mac": {
							Type:        schema.TypeString,
							Description: "MAC address of server interface to assign address to",
							Computed:    true,
						},
						"floating": {
							Type:        schema.TypeBool,
							Description: "Does the IP Address represents a floating IP Address",
							Computed:    true,
						},
						"zone": {
							Type:        schema.TypeString,
							Description: "Zone of address, required when assigning a detached floating IP address",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUpCloudIPAddressesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	ipAddresses, err := client.GetIPAddresses()

	if err != nil {
		diag.FromErr(err)
	}

	var values []map[string]interface{}

	for _, ipAddress := range ipAddresses.IPAddresses {

		value := map[string]interface{}{
			"access":       ipAddress.Access,
			"address":      ipAddress.Address,
			"family":       ipAddress.Family,
			"part_of_plan": ipAddress.PartOfPlan.Bool(),
			"ptr_record":   ipAddress.PTRRecord,
			"server":       ipAddress.ServerUUID,
			"mac":          ipAddress.MAC,
			"floating":     ipAddress.Floating.Bool(),
			"zone":         ipAddress.Zone,
		}

		values = append(values, value)
	}

	if err := d.Set("addresses", values); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(time.Now().UTC().String())

	return diags
}
