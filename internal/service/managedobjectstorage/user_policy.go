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

func ResourceManagedObjectStorageUserPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents an UpCloud Managed Object Storage user policy attachment.",
		CreateContext: resourceManagedObjectStorageUserPolicyCreate,
		ReadContext:   resourceManagedObjectStorageUserPolicyRead,
		DeleteContext: resourceManagedObjectStorageUserPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Description: "Policy ARN.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"name": {
				Description: "Policy name.",
				Required:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			"service_uuid": {
				Description: "Managed Object Storage service UUID.",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"username": {
				Description: "Username.",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceManagedObjectStorageUserPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	req := &request.AttachManagedObjectStorageUserPolicyRequest{
		Username:    d.Get("username").(string),
		ServiceUUID: d.Get("service_uuid").(string),
		Name:        d.Get("name").(string),
	}

	userPolicy, err := svc.AttachManagedObjectStorageUserPolicy(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(req.ServiceUUID, req.Username, userPolicy.Name))

	return setManagedObjectStorageUserPolicyData(d, userPolicy)
}

func resourceManagedObjectStorageUserPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)

	var serviceUUID, username, name string
	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &username, &name); err != nil {
		return diag.FromErr(err)
	}

	policies, err := svc.GetManagedObjectStorageUserPolicies(ctx, &request.GetManagedObjectStorageUserPoliciesRequest{
		ServiceUUID: serviceUUID,
		Username:    username,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	for _, policy := range policies {
		if policy.Name == name {
			return setManagedObjectStorageUserPolicyData(d, &policy)
		}
	}

	return nil
}

func resourceManagedObjectStorageUserPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)

	var serviceUUID, username, name string
	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &username, &name); err != nil {
		return diag.FromErr(err)
	}

	req := &request.DetachManagedObjectStorageUserPolicyRequest{
		ServiceUUID: serviceUUID,
		Username:    username,
		Name:        name,
	}

	if err := svc.DetachManagedObjectStorageUserPolicy(ctx, req); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setManagedObjectStorageUserPolicyData(d *schema.ResourceData, userPolicy *upcloud.ManagedObjectStorageUserPolicy) (diags diag.Diagnostics) {
	if err := d.Set("arn", userPolicy.Arn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", userPolicy.Name); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
