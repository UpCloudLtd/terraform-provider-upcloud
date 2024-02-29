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

func ResourceManagedObjectStorageUser() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents an UpCloud Managed Object Storage user. No relation to UpCloud API accounts.",
		CreateContext: resourceManagedObjectStorageUserCreate,
		ReadContext:   resourceManagedObjectStorageUserRead,
		DeleteContext: resourceManagedObjectStorageUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Description: "User ARN.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"created_at": {
				Description: "Creation time.",
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
				Description: "Custom usernames for accessing the object storage. No relation to UpCloud API accounts. See `upcloud_managed_object_storage_user_access_key` for managing access keys and `upcloud_managed_object_storage_user_policy` for managing policies.",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceManagedObjectStorageUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	req := &request.CreateManagedObjectStorageUserRequest{
		Username:    d.Get("username").(string),
		ServiceUUID: d.Get("service_uuid").(string),
	}

	accessKey, err := svc.CreateManagedObjectStorageUser(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(req.ServiceUUID, req.Username))

	return setManagedObjectStorageUserData(d, accessKey)
}

func resourceManagedObjectStorageUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)

	var serviceUUID, username, name string
	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &username, &name); err != nil {
		return diag.FromErr(err)
	}

	accessKey, err := svc.GetManagedObjectStorageUser(ctx, &request.GetManagedObjectStorageUserRequest{
		ServiceUUID: serviceUUID,
		Username:    username,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return setManagedObjectStorageUserData(d, accessKey)
}

func resourceManagedObjectStorageUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)

	var serviceUUID, username string
	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &username); err != nil {
		return diag.FromErr(err)
	}

	req := &request.DeleteManagedObjectStorageUserRequest{
		ServiceUUID: serviceUUID,
		Username:    username,
	}

	if err := svc.DeleteManagedObjectStorageUser(ctx, req); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setManagedObjectStorageUserData(d *schema.ResourceData, user *upcloud.ManagedObjectStorageUser) (diags diag.Diagnostics) {
	if err := d.Set("arn", user.ARN); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", user.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("username", user.Username); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
