package cloud

import (
	"context"
	"fmt"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	allFilter     string = "all"
	publicFilter  string = "public"
	privateFilter string = "private"
)

func DataSourceHosts() *schema.Resource {
	return &schema.Resource{
		Description: `Returns a list of available UpCloud hosts. 
		A host identifies the host server that virtual machines are run on. 
		Only hosts on private cloud to which the calling account has access to are available through this resource.`,
		ReadContext: dataSourceHostsRead,
		Schema: map[string]*schema.Schema{
			"hosts": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_id": {
							Type:        schema.TypeInt,
							Description: "The unique id of the host",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Free form text describing the host",
							Computed:    true,
						},
						"zone": {
							Type:        schema.TypeString,
							Description: "The zone the host is in, e.g. `de-fra1`. You can list available zones with `upctl zone list`.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceHostsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	hosts, err := client.GetHosts(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching hosts: %s", err))
	}

	var values []map[string]interface{}

	for _, host := range hosts.Hosts {
		value := map[string]interface{}{
			"host_id":     host.ID,
			"description": host.Description,
			"zone":        host.Zone,
		}

		values = append(values, value)
	}

	if err := d.Set("hosts", values); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(time.Now().UTC().String())

	return diags
}

func DataSourceZones() *schema.Resource {
	return &schema.Resource{
		Description: "Data-source is deprecated.",
		ReadContext: dataSourceZonesRead,
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

func dataSourceZonesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	zones, err := client.GetZones(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching zones: %s", err))
	}

	filterType, ok := d.GetOk("filter_type")

	if !ok {
		return diag.FromErr(fmt.Errorf("error getting filter_type: %s", err))
	}

	zoneIDs := utils.FilterZoneIDs(zones.Zones, func(zone upcloud.Zone) bool {
		switch filterType {
		case privateFilter:
			return zone.Public != upcloud.True
		case publicFilter:
			return zone.Public == upcloud.True
		default:
			return true
		}
	})

	if err := d.Set("zone_ids", zoneIDs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting zone_ids: %s", err))
	}
	d.SetId(time.Now().UTC().String())

	return diags
}
