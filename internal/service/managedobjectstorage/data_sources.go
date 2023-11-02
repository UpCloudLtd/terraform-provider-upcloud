package managedobjectstorage

import (
	"context"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceManagedObjectStorageRegions() *schema.Resource {
	return &schema.Resource{
		Description: `Returns a list of available Managed Object Storage regions.`,
		ReadContext: dataSourceHostsRead,
		Schema: map[string]*schema.Schema{
			"regions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the region.",
							Computed:    true,
						},
						"primary_zone": {
							Type:        schema.TypeString,
							Description: "Primary zone of the region.",
							Computed:    true,
						},
						"zones": {
							Type:        schema.TypeSet,
							Description: "List of zones in the region.",
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceHostsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	regions, err := svc.GetManagedObjectStorageRegions(ctx, &request.GetManagedObjectStorageRegionsRequest{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(time.Now().UTC().String())

	err = d.Set("regions", buildManagedObjectStorageRegions(regions))
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func buildManagedObjectStorageRegions(regions []upcloud.ManagedObjectStorageRegion) []map[string]interface{} {
	maps := make([]map[string]interface{}, 0)
	for _, region := range regions {
		zones := make([]string, 0)
		for _, zone := range region.Zones {
			zones = append(zones, zone.Name)
		}
		maps = append(maps, map[string]interface{}{
			"name":         region.Name,
			"primary_zone": region.PrimaryZone,
			"zones":        zones,
		})
	}

	return maps
}
