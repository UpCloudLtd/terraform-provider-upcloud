package managedobjectstorage

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceManagedObjectStorageUserAccessKey() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents an UpCloud Managed Object Storage user access key.",
		CreateContext: resourceManagedObjectStorageUserAccessKeyCreate,
		ReadContext:   resourceManagedObjectStorageUserAccessKeyRead,
		UpdateContext: resourceManagedObjectStorageUserAccessKeyUpdate,
		DeleteContext: resourceManagedObjectStorageUserAccessKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"access_key_id": {
				Description: "Access key id.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"created_at": {
				Description: "Creation time.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"last_used_at": {
				Description: "Last used.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"secret_access_key": {
				Description: "Secret access key.",
				Computed:    true,
				Sensitive:   true,
				Type:        schema.TypeString,
			},
			"service_uuid": {
				Description: "Managed Object Storage service UUID.",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"status": {
				Description:  "Status of the key. Valid values: `Active`|`Inactive`",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{string(upcloud.ManagedObjectStorageUserAccessKeyStatusActive), string(upcloud.ManagedObjectStorageUserAccessKeyStatusInactive)}, false),
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

func resourceManagedObjectStorageUserAccessKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	req := &request.CreateManagedObjectStorageUserAccessKeyRequest{
		Username:    d.Get("username").(string),
		ServiceUUID: d.Get("service_uuid").(string),
	}

	accessKey, err := svc.CreateManagedObjectStorageUserAccessKey(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(req.ServiceUUID, req.Username, accessKey.AccessKeyID))

	status := upcloud.ManagedObjectStorageUserAccessKeyStatus(d.Get("status").(string))
	if status != accessKey.Status {
		accessKey, err = svc.ModifyManagedObjectStorageUserAccessKey(ctx, &request.ModifyManagedObjectStorageUserAccessKeyRequest{
			Username:    d.Get("username").(string),
			ServiceUUID: d.Get("service_uuid").(string),
			AccessKeyID: accessKey.AccessKeyID,
			Status:      status,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return setManagedObjectStorageUserAccessKeyData(d, accessKey)
}

func resourceManagedObjectStorageUserAccessKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)

	var serviceUUID, username, id string
	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &username, &id); err != nil {
		return diag.FromErr(err)
	}

	accessKey, err := svc.GetManagedObjectStorageUserAccessKey(ctx, &request.GetManagedObjectStorageUserAccessKeyRequest{
		ServiceUUID: serviceUUID,
		Username:    username,
		AccessKeyID: id,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return setManagedObjectStorageUserAccessKeyData(d, accessKey)
}

func resourceManagedObjectStorageUserAccessKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	if !d.HasChange("status") {
		return nil
	}

	svc := meta.(*service.Service)

	var serviceUUID, username, id string
	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &username, &id); err != nil {
		return diag.FromErr(err)
	}

	req := &request.ModifyManagedObjectStorageUserAccessKeyRequest{
		Status:      upcloud.ManagedObjectStorageUserAccessKeyStatus(d.Get("status").(string)),
		ServiceUUID: serviceUUID,
		Username:    username,
		AccessKeyID: id,
	}

	accessKey, err := svc.ModifyManagedObjectStorageUserAccessKey(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return setManagedObjectStorageUserAccessKeyData(d, accessKey)
}

func resourceManagedObjectStorageUserAccessKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)

	var serviceUUID, username, id string
	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &username, &id); err != nil {
		return diag.FromErr(err)
	}

	req := &request.DeleteManagedObjectStorageUserAccessKeyRequest{
		ServiceUUID: serviceUUID,
		Username:    username,
		AccessKeyID: id,
	}

	if err := svc.DeleteManagedObjectStorageUserAccessKey(ctx, req); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setManagedObjectStorageUserAccessKeyData(d *schema.ResourceData, accessKey *upcloud.ManagedObjectStorageUserAccessKey) (diags diag.Diagnostics) {
	if err := d.Set("access_key_id", accessKey.AccessKeyID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", accessKey.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", accessKey.Status); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("last_used_at", accessKey.LastUsedAt.String()); err != nil {
		return diag.FromErr(err)
	}
	if accessKey.SecretAccessKey != nil {
		if err := d.Set("secret_access_key", accessKey.SecretAccessKey); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
