package managedobjectstorage

import (
	"context"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
