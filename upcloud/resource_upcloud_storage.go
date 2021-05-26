package upcloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/server"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/storage"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUpCloudStorage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUpCloudStorageCreate,
		ReadContext:   resourceUpCloudStorageRead,
		UpdateContext: resourceUpCloudStorageUpdate,
		DeleteContext: resourceUpCloudStorageDelete,
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
			"backup_rule": storage.BackupRuleSchema(),
		},
	}
}

func resourceUpCloudStorageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

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
		diags = createStorage(client, size, tier, title, zone, d)
	} else {
		diags = cloneStorage(client, size, tier, title, zone, d)
	}
	if diags.HasError() {
		return diags
	}

	diags = append(diags, resourceUpCloudStorageRead(ctx, d, meta)...)

	return diags
}

func cloneStorage(
	client *service.Service,
	size int,
	tier string,
	title string,
	zone string,
	d *schema.ResourceData) diag.Diagnostics {
	cloneStorageRequest := request.CloneStorageRequest{
		Zone:  zone,
		Tier:  tier,
		Title: title,
	}

	if v, ok := d.GetOk("clone"); ok {
		block := v.(*schema.Set).List()[0].(map[string]interface{})
		cloneStorageRequest.UUID = block["id"].(string)
	}

	originalStorageDevice, err := client.WaitForStorageState(&request.WaitForStorageStateRequest{
		UUID:         cloneStorageRequest.UUID,
		DesiredState: upcloud.StorageStateOnline,
		Timeout:      15 * time.Minute,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if originalStorageDevice.Size > size {
		return diag.Errorf("cloned storage device should be at least the same size as the original one")
	}

	storage, err := client.CloneStorage(&cloneStorageRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	storage, err = client.WaitForStorageState(&request.WaitForStorageStateRequest{
		UUID:         storage.UUID,
		DesiredState: upcloud.StorageStateOnline,
		Timeout:      15 * time.Minute,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// If the storage specified does not match the cloned storage, modify it so that it does.
	if storage.Size != size {
		storage, err := client.ModifyStorage(&request.ModifyStorageRequest{
			UUID: storage.UUID,
			Size: size,
		})
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = client.WaitForStorageState(&request.WaitForStorageStateRequest{
			UUID:         storage.UUID,
			DesiredState: upcloud.StorageStateOnline,
			Timeout:      15 * time.Minute,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(storage.UUID)

	return nil
}

func createStorage(
	client *service.Service,
	size int,
	tier string,
	title string,
	zone string,
	d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	createStorageRequest := request.CreateStorageRequest{
		Size:  size,
		Tier:  tier,
		Title: title,
		Zone:  zone,
	}

	var importReq *request.CreateStorageImportRequest
	if v, ok := d.GetOk("import"); ok {
		importReq = &request.CreateStorageImportRequest{}
		importBlock := v.(*schema.Set).List()[0].(map[string]interface{})
		if impV, ok := importBlock["source"]; ok {
			importReq.Source = impV.(string)
		}
		if impV, ok := importBlock["source_location"]; ok {
			importReq.SourceLocation = impV.(string)
		}
	}

	if v, ok := d.GetOk("backup_rule.0"); ok {
		createStorageRequest.BackupRule = storage.BackupRule(v.(map[string]interface{}))
	}

	storage, err := client.CreateStorage(&createStorageRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for storage to enter the 'online' state. For a fresh storage device
	// this is pretty quick.
	_, err = client.WaitForStorageState(&request.WaitForStorageStateRequest{
		UUID:         storage.UUID,
		DesiredState: upcloud.StorageStateOnline,
		Timeout:      15 * time.Minute,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if importReq != nil {
		importReq.StorageUUID = storage.UUID
		_, err := client.CreateStorageImport(importReq)
		if err != nil {
			return diagAndTidy(client, storage.UUID, err)
		}

		_, err = client.WaitForStorageImportCompletion(&request.WaitForStorageImportCompletionRequest{
			StorageUUID: storage.UUID,
			Timeout:     15 * time.Minute,
		})
		if err != nil {
			return diagAndTidy(client, storage.UUID, err)
		}

		// Imported storage will enter a 'syncing' state for a while. Storage in this
		// state can be used by a server so we will wait for that to allow progress.
		_, err = client.WaitForStorageState(&request.WaitForStorageStateRequest{
			UUID:         storage.UUID,
			DesiredState: upcloud.StorageStateSyncing,
			Timeout:      15 * time.Minute,
		})
		if err != nil {
			return diagAndTidy(client, storage.UUID, err)
		}
	}

	d.SetId(storage.UUID)

	return diags
}

func diagAndTidy(client *service.Service, storageUUID string, err error) diag.Diagnostics {
	_, waitErr := client.WaitForStorageState(&request.WaitForStorageStateRequest{
		UUID:         storageUUID,
		DesiredState: upcloud.StorageStateOnline,
		Timeout:      15 * time.Minute,
	})
	if waitErr != nil {
		return diag.Errorf("wait for storage after import error: %s", waitErr.Error())
	}

	delErr := client.DeleteStorage(&request.DeleteStorageRequest{
		UUID: storageUUID,
	})
	if delErr != nil {
		return diag.Errorf("delete storage after import error: %s", delErr.Error())
	}
	return diag.Errorf("storage import error: %s", err.Error())
}

func resourceUpCloudStorageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	r := &request.GetStorageDetailsRequest{
		UUID: d.Id(),
	}
	storage, err := client.GetStorageDetails(r)

	if err != nil {
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

	if storage.BackupRule != nil && storage.BackupRule.Retention > 0 {
		backupRule := []interface{}{
			map[string]interface{}{
				"interval":  storage.BackupRule.Interval,
				"time":      storage.BackupRule.Time,
				"retention": storage.BackupRule.Retention,
			},
		}

		if err := d.Set("backup_rule", backupRule); err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := d.GetOk("import"); ok {
		configImportBlock := v.(*schema.Set).List()[0].(map[string]interface{})

		_, err := client.WaitForStorageImportCompletion(&request.WaitForStorageImportCompletionRequest{
			StorageUUID: d.Id(),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		importDetails, err := client.GetStorageImportDetails(&request.GetStorageImportDetailsRequest{
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

func resourceUpCloudStorageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	_, err := client.WaitForStorageState(&request.WaitForStorageStateRequest{
		UUID:         d.Id(),
		DesiredState: upcloud.StorageStateOnline,
		Timeout:      15 * time.Minute,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	req := request.ModifyStorageRequest{
		UUID:       d.Id(),
		Size:       d.Get("size").(int),
		Title:      d.Get("title").(string),
		BackupRule: storage.BackupRule(d.Get("backup_rule.0").(map[string]interface{})),
	}

	storageDetails, err := client.GetStorageDetails(&request.GetStorageDetailsRequest{
		UUID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	// need to shut down server if resizing
	if len(storageDetails.ServerUUIDs) > 0 && d.HasChange("size") {
		err := server.VerifyServerStopped(request.StopServerRequest{UUID: storageDetails.ServerUUIDs[0]}, meta)
		if err != nil {
			return diag.FromErr(err)
		}
		if _, err := utils.WithRetry(func() (interface{}, error) { return client.ModifyStorage(&req) }, 20, time.Second*5); err != nil {
			return diag.FromErr(err)
		}

		// No need to pass host explicitly here, as the server will be started on old host by default (for private clouds)
		if err = server.VerifyServerStarted(request.StartServerRequest{UUID: storageDetails.ServerUUIDs[0]}, meta); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if _, err := client.ModifyStorage(&req); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceUpCloudStorageRead(ctx, d, meta)
}

func resourceUpCloudStorageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	// Wait for storage to enter 'online' state as storage devices can only
	// be deleted in this state.
	_, err := client.WaitForStorageState(&request.WaitForStorageStateRequest{
		UUID:         d.Id(),
		DesiredState: upcloud.StorageStateOnline,
		Timeout:      15 * time.Minute,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// fetch storage details for checking that the storage can be deleted
	storageDetails, err := client.GetStorageDetails(&request.GetStorageDetailsRequest{
		UUID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if len(storageDetails.ServerUUIDs) > 0 {
		serverUUID := storageDetails.ServerUUIDs[0]
		// Get server details for retrieving the address that is to be used when detaching the storage
		serverDetails, err := client.GetServerDetails(&request.GetServerDetailsRequest{
			UUID: serverUUID,
		})
		if err != nil {
			return diag.FromErr(err)
		}

		if storageDevice := serverDetails.StorageDevice(d.Id()); storageDevice != nil {
			// ide devices can only be detached from stopped servers
			if strings.HasPrefix(storageDevice.Address, "ide") {
				err = server.VerifyServerStopped(request.StopServerRequest{UUID: serverUUID}, meta)
				if err != nil {
					return diag.FromErr(err)
				}
			}

			_, err = utils.WithRetry(func() (interface{}, error) {
				return client.DetachStorage(&request.DetachStorageRequest{ServerUUID: serverUUID, Address: storageDevice.Address})
			}, 20, time.Second*3)
			if err != nil {
				return diag.FromErr(err)
			}

			if strings.HasPrefix(storageDevice.Address, "ide") && serverDetails.State != upcloud.ServerStateStopped {
				// No need to pass host explicitly here, as the server will be started on old host by default (for private clouds)
				if err = server.VerifyServerStarted(request.StartServerRequest{UUID: serverUUID}, meta); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	deleteStorageRequest := &request.DeleteStorageRequest{
		UUID: d.Id(),
	}
	err = client.DeleteStorage(deleteStorageRequest)

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
