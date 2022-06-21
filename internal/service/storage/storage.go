package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
)

const storageNotFoundErrorCode string = "STORAGE_NOT_FOUND"

func ResourceStorage() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages UpCloud storage block devices.",
		CreateContext: resourceStorageCreate,
		ReadContext:   resourceStorageRead,
		UpdateContext: resourceStorageUpdate,
		DeleteContext: resourceStorageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"size": {
				Description:  "The size of the storage in gigabytes",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(10, 2048),
			},
			"tier": {
				Description:  "The storage tier to use",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"hdd", "maxiops"}, false),
			},
			"title": {
				Description:  "A short, informative description",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"zone": {
				Description: "The zone in which the storage will be created",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"clone": {
				Description:   "Block defining another storage/template to clone to storage",
				Type:          schema.TypeSet,
				MaxItems:      1,
				MinItems:      0,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"import"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The unique identifier of the storage/template to clone",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"import": {
				Description:   "Block defining external data to import to storage",
				Type:          schema.TypeSet,
				MaxItems:      1,
				MinItems:      0,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"clone"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Description: "The mode of the import task. One of `http_import` or `direct_upload`.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							ValidateDiagFunc: func(v interface{}, _ cty.Path) diag.Diagnostics {
								switch v.(string) {
								case upcloud.StorageImportSourceDirectUpload, upcloud.StorageImportSourceHTTPImport:
									return nil
								default:
									return diag.Diagnostics{diag.Diagnostic{
										Severity: diag.Error,
										Summary:  "'source' value incorrect",
										Detail: fmt.Sprintf("'source' must be '%s' or '%s'",
											upcloud.StorageImportSourceDirectUpload,
											upcloud.StorageImportSourceHTTPImport),
									}}
								}
							},
						},
						"source_location": {
							Description: "The location of the file to import. For `http_import` an accessible URL for `direct_upload` a local file.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"source_hash": {
							Description: "For `direct_upload`; an optional hash of the file to upload.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
						"sha256sum": {
							Description: "sha256 sum of the imported data",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"written_bytes": {
							Description: "Number of bytes imported",
							Type:        schema.TypeInt,
							Computed:    true,
						},
					},
				},
			},
			"backup_rule": BackupRuleSchema(),
			"filesystem_autoresize": {
				Description: `If set to true, provider will attempt to resize partition and filesystem when the size of the storage changes.
				Please note that before the resize attempt is made, backup of the storage will be taken. If the resize attempt fails, the backup will be used
				to restore the storage and then deleted. If the resize attempt succeeds, backup will be kept (unless delete_autoresize_backup option is set to true).
				Taking and keeping backups incure costs.`,
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"delete_autoresize_backup": {
				Description: "If set to true, the backup taken before the partition and filesystem resize attempt will be deleted immediately after success.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceStorageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.ServiceContext)

	var diags diag.Diagnostics

	var size int
	var tier, title, zone string

	if v, ok := d.GetOk("size"); ok {
		size = v.(int)
	}
	if v, ok := d.GetOk("tier"); ok {
		tier = v.(string)
	}
	if v, ok := d.GetOk("title"); ok {
		title = v.(string)
	}
	if v, ok := d.GetOk("zone"); ok {
		zone = v.(string)
	}

	if _, ok := d.GetOk("clone"); !ok {
		// There is not 'clone' block so do the
		// create storage logic including importing
		// external data.
		diags = createStorage(ctx, client, size, tier, title, zone, d)
	} else {
		diags = cloneStorage(ctx, client, size, tier, title, zone, d)
	}
	if diags.HasError() {
		return diags
	}

	if err := d.Set("filesystem_autoresize", d.Get("filesystem_autoresize")); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("delete_autoresize_backup", d.Get("delete_autoresize_backup")); err != nil {
		return diag.FromErr(err)
	}

	diags = append(diags, resourceStorageRead(ctx, d, meta)...)

	return diags
}

func resourceStorageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.ServiceContext)

	var diags diag.Diagnostics

	r := &request.GetStorageDetailsRequest{
		UUID: d.Id(),
	}
	storage, err := client.GetStorageDetails(ctx, r)

	if err != nil {
		if svcErr, ok := err.(*upcloud.Error); ok && svcErr.ErrorCode == storageNotFoundErrorCode {
			diags = append(diags, utils.DiagBindingRemovedWarningFromUpcloudErr(svcErr, d.Get("title").(string)))
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("filesystem_autoresize", d.Get("filesystem_autoresize")); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("delete_autoresize_backup", d.Get("delete_autoresize_backup")); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("size", storage.Size); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("title", storage.Title); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("tier", storage.Tier); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("zone", storage.Zone); err != nil {
		return diag.FromErr(err)
	}

	simpleBackupEnabled, err := isStorageSimpleBackupEnabled(ctx, client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	// Set backup_rule only if simple backup is not in use.
	// This is to avoid conflicting backup rules when using simple_backups with server that storage is attached to
	if !simpleBackupEnabled {
		if storage.BackupRule != nil && storage.BackupRule.Retention > 0 {
			backupRule := []interface{}{
				map[string]interface{}{
					"interval":  storage.BackupRule.Interval,
					"time":      storage.BackupRule.Time,
					"retention": storage.BackupRule.Retention,
				},
			}

			if err := d.Set("backup_rule", backupRule); err != nil {
				tflog.Debug(ctx, "error on set simple backup")
				return diag.FromErr(err)
			}
		}
	}

	if v, ok := d.GetOk("import"); ok {
		configImportBlock := v.(*schema.Set).List()[0].(map[string]interface{})

		_, err := client.WaitForStorageImportCompletion(ctx, &request.WaitForStorageImportCompletionRequest{
			StorageUUID: d.Id(),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		importDetails, err := client.GetStorageImportDetails(ctx, &request.GetStorageImportDetailsRequest{
			UUID: d.Id(),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		importBlock := []interface{}{
			map[string]interface{}{
				"sha256sum":       importDetails.SHA256Sum,
				"written_bytes":   importDetails.WrittenBytes,
				"source":          configImportBlock["source"],
				"source_location": configImportBlock["source_location"],
				"source_hash":     configImportBlock["source_hash"],
			},
		}

		_ = d.Set("import", importBlock)
	}

	return diags
}

func resourceStorageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.ServiceContext)
	diags := diag.Diagnostics{}

	_, err := client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         d.Id(),
		DesiredState: upcloud.StorageStateOnline,
		Timeout:      15 * time.Minute,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	req := request.ModifyStorageRequest{
		UUID:  d.Id(),
		Size:  d.Get("size").(int),
		Title: d.Get("title").(string),
	}

	if d.HasChange("backup_rule") {
		if br, ok := d.GetOk("backup_rule.0"); ok {
			backupRule := BackupRule(br.(map[string]interface{}))
			if backupRule.Interval == "" {
				req.BackupRule = &upcloud.BackupRule{}
			} else {
				req.BackupRule = backupRule
			}
		}
	}

	storageDetails, err := client.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{
		UUID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	// need to shut down server if resizing
	if len(storageDetails.ServerUUIDs) > 0 && d.HasChange("size") {
		err := utils.VerifyServerStopped(ctx, request.StopServerRequest{UUID: storageDetails.ServerUUIDs[0]}, meta)
		if err != nil {
			return diag.FromErr(err)
		}
		if _, err := utils.WithRetry(func() (interface{}, error) { return client.ModifyStorage(ctx, &req) }, 20, time.Second*5); err != nil {
			return diag.FromErr(err)
		}

		if d.Get("filesystem_autoresize").(bool) {
			diags = append(diags, ResizeStoragePartitionAndFs(
				ctx,
				client,
				storageDetails.UUID,
				storageDetails.Title,
				d.Get("delete_autoresize_backup").(bool),
			)...)
		}

		// No need to pass host explicitly here, as the server will be started on old host by default (for private clouds)
		if err = utils.VerifyServerStarted(ctx, request.StartServerRequest{UUID: storageDetails.ServerUUIDs[0]}, meta); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if _, err := client.ModifyStorage(ctx, &req); err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("size") && d.Get("filesystem_autoresize").(bool) {
			diags = append(diags, ResizeStoragePartitionAndFs(
				ctx,
				client,
				storageDetails.UUID,
				storageDetails.Title,
				d.Get("delete_autoresize_backup").(bool),
			)...)
		}
	}

	diags = append(diags, resourceStorageRead(ctx, d, meta)...)

	return diags
}

func resourceStorageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.ServiceContext)

	var diags diag.Diagnostics

	// Wait for storage to enter 'online' state as storage devices can only
	// be deleted in this state.
	_, err := client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         d.Id(),
		DesiredState: upcloud.StorageStateOnline,
		Timeout:      15 * time.Minute,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// fetch storage details for checking that the storage can be deleted
	storageDetails, err := client.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{
		UUID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if len(storageDetails.ServerUUIDs) > 0 {
		serverUUID := storageDetails.ServerUUIDs[0]
		// Get server details for retrieving the address that is to be used when detaching the storage
		serverDetails, err := client.GetServerDetails(ctx, &request.GetServerDetailsRequest{
			UUID: serverUUID,
		})
		if err != nil {
			return diag.FromErr(err)
		}

		if storageDevice := serverDetails.StorageDevice(d.Id()); storageDevice != nil {
			// ide devices can only be detached from stopped servers
			if strings.HasPrefix(storageDevice.Address, "ide") {
				err = utils.VerifyServerStopped(ctx, request.StopServerRequest{UUID: serverUUID}, meta)
				if err != nil {
					return diag.FromErr(err)
				}
			}

			_, err = utils.WithRetry(func() (interface{}, error) {
				return client.DetachStorage(ctx, &request.DetachStorageRequest{ServerUUID: serverUUID, Address: storageDevice.Address})
			}, 20, time.Second*3)
			if err != nil {
				return diag.FromErr(err)
			}

			if strings.HasPrefix(storageDevice.Address, "ide") && serverDetails.State != upcloud.ServerStateStopped {
				// No need to pass host explicitly here, as the server will be started on old host by default (for private clouds)
				if err = utils.VerifyServerStarted(ctx, request.StartServerRequest{UUID: serverUUID}, meta); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	deleteStorageRequest := &request.DeleteStorageRequest{
		UUID: d.Id(),
	}
	err = client.DeleteStorage(ctx, deleteStorageRequest)

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
