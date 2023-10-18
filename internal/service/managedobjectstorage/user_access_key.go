package managedobjectstorage

import (
	"context"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
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
			"enabled": {
				Description: "Enabled or not.",
				Required:    true,
				Type:        schema.TypeBool,
			},
			"last_used_at": {
				Description: "Last used.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"name": {
				Description: "Access key name. Must be unique within the user.",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), ""),
				),
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
			"updated_at": {
				Description: "Update time.",
				Computed:    true,
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

func resourceManagedObjectStorageUserAccessKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	req := &request.CreateManagedObjectStorageUserAccessKeyRequest{
		Username:    d.Get("username").(string),
		ServiceUUID: d.Get("service_uuid").(string),
		Name:        d.Get("name").(string),
		Enabled:     d.Get("enabled").(bool),
	}

	accessKey, err := svc.CreateManagedObjectStorageUserAccessKey(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(req.ServiceUUID, req.Username, req.Name))

	return setManagedObjectStorageUserAccessKeyData(d, accessKey)
}

func resourceManagedObjectStorageUserAccessKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)

	var serviceUUID, username, name string
	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &username, &name); err != nil {
		return diag.FromErr(err)
	}

	accessKey, err := svc.GetManagedObjectStorageUserAccessKey(ctx, &request.GetManagedObjectStorageUserAccessKeyRequest{
		ServiceUUID: serviceUUID,
		Username:    username,
		Name:        name,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return setManagedObjectStorageUserAccessKeyData(d, accessKey)
}

func resourceManagedObjectStorageUserAccessKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	if !d.HasChange("enabled") {
		return nil
	}

	svc := meta.(*service.Service)

	var serviceUUID, username, name string
	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &username, &name); err != nil {
		return diag.FromErr(err)
	}

	req := &request.ModifyManagedObjectStorageUserAccessKeyRequest{
		Enabled:     d.Get("enabled").(bool),
		ServiceUUID: serviceUUID,
		Username:    username,
		Name:        name,
	}

	accessKey, err := svc.ModifyManagedObjectStorageUserAccessKey(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return setManagedObjectStorageUserAccessKeyData(d, accessKey)
}

func resourceManagedObjectStorageUserAccessKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)

	var serviceUUID, username, name string
	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &username, &name); err != nil {
		return diag.FromErr(err)
	}

	req := &request.DeleteManagedObjectStorageUserAccessKeyRequest{
		ServiceUUID: serviceUUID,
		Username:    username,
		Name:        name,
	}

	if err := svc.DeleteManagedObjectStorageUserAccessKey(ctx, req); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setManagedObjectStorageUserAccessKeyData(d *schema.ResourceData, accessKey *upcloud.ManagedObjectStorageUserAccessKey) (diags diag.Diagnostics) {
	if err := d.Set("access_key_id", accessKey.AccessKeyId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", accessKey.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", accessKey.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("last_used_at", accessKey.LastUsedAt.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", accessKey.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", accessKey.UpdatedAt.String()); err != nil {
		return diag.FromErr(err)
	}
	if accessKey.SecretAccessKey != nil {
		if err := d.Set("secret_access_key", accessKey.SecretAccessKey); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
