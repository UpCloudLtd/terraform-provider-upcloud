package upcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	allFilter     = "all"
	publicFilter  = "public"
	privateFilter = "private"
)

func dataSourceUpCloudZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUpCloudZonesRead,
		Schema: map[string]*schema.Schema{
			"zone_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      allFilter,
				ValidateFunc: validation.StringInSlice([]string{allFilter, publicFilter, privateFilter}, false),
			},
		},
	}
}

func dataSourceUpCloudZonesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	zones, err := client.GetZones()

	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching zones: %s", err))
	}

	filterType, ok := d.GetOk("filter_type")

	if !ok {
		return diag.FromErr(fmt.Errorf("error getting filter_type: %s", err))
	}

	zoneIds := utils.FilterZoneIds(zones.Zones, func(zone upcloud.Zone) bool {
		switch filterType {
		case privateFilter:
			return zone.Public != upcloud.True
		case publicFilter:
			return zone.Public == upcloud.True
		default:
			return true
		}
	})

	if err := d.Set("zone_ids", zoneIds); err != nil {
		return diag.FromErr(fmt.Errorf("error setting zone_ids: %s", err))
	}
	d.SetId(time.Now().UTC().String())

	return diags
}
