package managedobjectstorage

import (
	"context"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceManagedObjectStorageRegions() *schema.Resource {
	return &schema.Resource{
		Description: `Returns a list of available Managed Object Storage regions.`,
		ReadContext: dataSourceRegionsRead,
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

func dataSourceRegionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
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

func DataSourceManagedObjectStoragePolicies() *schema.Resource {
	return &schema.Resource{
		Description: "Policies available for a Managed Object Storage resource. See `managed_object_storage_user_policy` for attaching to a user.",
		ReadContext: dataSourcePoliciesRead,
		Schema: map[string]*schema.Schema{
			"policies": {
				Description: "Policies.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: schemaPolicy(),
				},
			},
			"service_uuid": {
				Description: "Service UUID.",
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func dataSourcePoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	serviceUUID := d.Get("service_uuid").(string)

	policies, err := svc.GetManagedObjectStoragePolicies(ctx, &request.GetManagedObjectStoragePoliciesRequest{ServiceUUID: serviceUUID})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serviceUUID)

	return setManagedObjectStoragePoliciesData(d, policies)
}

func setManagedObjectStoragePoliciesData(d *schema.ResourceData, policies []upcloud.ManagedObjectStoragePolicy) (diags diag.Diagnostics) {
	policyMaps := make([]map[string]interface{}, 0)
	for _, policy := range policies {
		policyMaps = append(policyMaps, map[string]interface{}{
			"arn":                policy.ARN,
			"attachment_count":   policy.AttachmentCount,
			"created_at":         policy.CreatedAt.String(),
			"default_version_id": policy.DefaultVersionID,
			"description":        policy.Description,
			"document":           policy.Document,
			"name":               policy.Name,
			"system":             policy.System,
			"updated_at":         policy.UpdatedAt.String(),
		})
	}
	if err := d.Set("policies", policyMaps); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
