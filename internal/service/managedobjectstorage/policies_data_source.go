package managedobjectstorage

import (
	"context"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaPolicy() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"arn": {
			Description: "Policy ARN.",
			Computed:    true,
			Type:        schema.TypeString,
		},
		"attachment_count": {
			Description: "Attachment count.",
			Computed:    true,
			Type:        schema.TypeInt,
		},
		"created_at": {
			Description: "Creation time.",
			Computed:    true,
			Type:        schema.TypeString,
		},
		"default_version_id": {
			Description: "Default version id.",
			Computed:    true,
			Type:        schema.TypeString,
		},
		"description": {
			Description: "Description of the policy.",
			Optional:    true,
			ForceNew:    true,
			Type:        schema.TypeString,
		},
		"document": {
			Description: "Policy document, URL-encoded compliant with RFC 3986.",
			Required:    true,
			ForceNew:    true,
			Type:        schema.TypeString,
		},
		"name": {
			Description: "Policy name.",
			Required:    true,
			ForceNew:    true,
			Type:        schema.TypeString,
		},
		"service_uuid": {
			Description: "Managed Object Storage service UUID.",
			Required:    true,
			ForceNew:    true,
			Type:        schema.TypeString,
		},
		"system": {
			Description: "Defines whether the policy was set up by the system.",
			Computed:    true,
			Type:        schema.TypeBool,
		},
		"updated_at": {
			Description: "Update time.",
			Computed:    true,
			Type:        schema.TypeString,
		},
	}
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
