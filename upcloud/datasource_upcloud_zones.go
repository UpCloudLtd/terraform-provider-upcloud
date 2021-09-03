package upcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
)

const (
	AllFilter     = "all"
	PublicFilter  = "public"
	PrivateFilter = "private"
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
				Default:      AllFilter,
				ValidateFunc: validation.StringInSlice([]string{AllFilter, PublicFilter, PrivateFilter}, false),
			},
		},
	}
}

func dataSourceUpCloudZonesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	zones, err := client.GetZones()
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching zones: %w", err))
	}

	filterType, ok := d.GetOk("filter_type")

	if !ok {
		return diag.FromErr(fmt.Errorf("error getting filter_type: %w", err))
	}

	zoneIds := utils.FilterZoneIds(zones.Zones, func(zone upcloud.Zone) bool {
		switch filterType {
		case PrivateFilter:
			return zone.Public != upcloud.True
		case PublicFilter:
			return zone.Public == upcloud.True
		default:
			return true
		}
	})

	if err := d.Set("zone_ids", zoneIds); err != nil {
		return diag.FromErr(fmt.Errorf("error setting zone_ids: %w", err))
	}
	d.SetId(time.Now().UTC().String())

	return diags
}
