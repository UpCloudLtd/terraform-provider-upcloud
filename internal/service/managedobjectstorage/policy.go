package managedobjectstorage

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceManagedObjectStoragePolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents an UpCloud Managed Object Storage policy.",
		CreateContext: resourceManagedObjectStoragePolicyCreate,
		ReadContext:   resourceManagedObjectStoragePolicyRead,
		DeleteContext: resourceManagedObjectStoragePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: schemaPolicy(),
	}
}

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

func resourceManagedObjectStoragePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	req := &request.CreateManagedObjectStoragePolicyRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Document:    d.Get("document").(string),
		ServiceUUID: d.Get("service_uuid").(string),
	}

	policy, err := svc.CreateManagedObjectStoragePolicy(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(req.ServiceUUID, req.Name))

	return setManagedObjectStoragePolicyData(d, policy)
}

func resourceManagedObjectStoragePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)

	var serviceUUID, name string
	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &name); err != nil {
		return diag.FromErr(err)
	}

	policy, err := svc.GetManagedObjectStoragePolicy(ctx, &request.GetManagedObjectStoragePolicyRequest{
		ServiceUUID: serviceUUID,
		Name:        name,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return setManagedObjectStoragePolicyData(d, policy)
}

func resourceManagedObjectStoragePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)

	var serviceUUID, name string
	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &name); err != nil {
		return diag.FromErr(err)
	}

	req := &request.DeleteManagedObjectStoragePolicyRequest{
		ServiceUUID: serviceUUID,
		Name:        name,
	}

	if err := svc.DeleteManagedObjectStoragePolicy(ctx, req); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setManagedObjectStoragePolicyData(d *schema.ResourceData, policy *upcloud.ManagedObjectStoragePolicy) (diags diag.Diagnostics) {
	if err := d.Set("arn", policy.ARN); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("attachment_count", policy.AttachmentCount); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", policy.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_version_id", policy.DefaultVersionID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("system", policy.System); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", policy.UpdatedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
