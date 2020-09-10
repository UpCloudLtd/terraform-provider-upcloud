package upcloud

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUpCloudZone() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceUpCloudZoneRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Unique lablel for the zone",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Meaningful text describing the zone",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"public": {
				Description: "Indicates whether the zone is public",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func resourceUpCloudZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*service.Service)

	var diags diag.Diagnostics

	zones, err := client.GetZones()

	if err != nil {
		return diag.FromErr(fmt.Errorf("Error fetching Zones: %s", err))
	}

	var locatedZone upcloud.Zone

	if v, ok := d.GetOk("name"); ok {
		zoneId := v.(string)
		zones := FilterZones(zones.Zones, func(zone upcloud.Zone) bool {
			return zone.ID == zoneId
		})

		if len(zones) > 1 {
			return diag.FromErr(fmt.Errorf("Error multiple Zones located: %s", err))
		}

		locatedZone = zones[0]
	}

	if err := d.Set("description", locatedZone.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting zone description: %s", err))
	}

	if err := d.Set("public", locatedZone.Public.Bool()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting zone public state: %s", err))
	}

	d.SetId(locatedZone.ID)

	return diags
}
