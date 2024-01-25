package storage

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResizeStoragePartitionAndFs(ctx context.Context, client *service.Service, UUID, title string, deleteBackup bool) diag.Diagnostics {
	diags := diag.Diagnostics{}

	backup, err := client.ResizeStorageFilesystem(ctx, &request.ResizeStorageFilesystemRequest{
		UUID: UUID,
	})
	if err != nil {
		summary := fmt.Sprintf(
			"Failed to resize partition and filesystem for storage %s(%s)",
			UUID,
			title,
		)

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  summary,
			Detail:   err.Error(),
		})
	}

	if err == nil && deleteBackup {
		err = client.DeleteStorage(ctx, &request.DeleteStorageRequest{
			UUID: backup.UUID,
		})

		if err != nil {
			summary := fmt.Sprintf(
				"Failed to delete the backup of storage %s(%s) after the partition and filesystem resize; you will need to delete the backup manually",
				UUID,
				title,
			)

			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  summary,
				Detail:   err.Error(),
			})
		}
	}

	return diags
}

func cloneStorage(
	ctx context.Context,
	client *service.Service,
	encrypted bool,
	size int,
	tier string,
	title string,
	zone string,
	d *schema.ResourceData,
) diag.Diagnostics {
	cloneStorageRequest := request.CloneStorageRequest{
		Encrypted: upcloud.FromBool(encrypted),
		Zone:      zone,
		Tier:      tier,
		Title:     title,
	}

	if v, ok := d.GetOk("clone"); ok {
		block := v.(*schema.Set).List()[0].(map[string]interface{})
		cloneStorageRequest.UUID = block["id"].(string)
	}

	originalStorageDevice, err := client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         cloneStorageRequest.UUID,
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if originalStorageDevice.Size > size {
		return diag.Errorf("cloned storage device should be at least the same size as the original one")
	}

	storage, err := client.CloneStorage(ctx, &cloneStorageRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	storage, err = client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         storage.UUID,
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// If the storage specified does not match the cloned storage, modify it so that it does.
	if storage.Size != size {
		storage, err := client.ModifyStorage(ctx, &request.ModifyStorageRequest{
			UUID: storage.UUID,
			Size: size,
		})
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
			UUID:         storage.UUID,
			DesiredState: upcloud.StorageStateOnline,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(storage.UUID)

	return nil
}

func createStorage(
	ctx context.Context,
	client *service.Service,
	encrypted bool,
	size int,
	tier string,
	title string,
	zone string,
	d *schema.ResourceData,
) diag.Diagnostics {
	var diags diag.Diagnostics

	createStorageRequest := request.CreateStorageRequest{
		Encrypted: upcloud.FromBool(encrypted),
		Size:      size,
		Tier:      tier,
		Title:     title,
		Zone:      zone,
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
		createStorageRequest.BackupRule = BackupRule(v.(map[string]interface{}))
	}

	storage, err := client.CreateStorage(ctx, &createStorageRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for storage to enter the 'online' state. For a fresh storage device
	// this is pretty quick.
	_, err = client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         storage.UUID,
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if importReq != nil {
		importReq.StorageUUID = storage.UUID
		_, err := client.CreateStorageImport(ctx, importReq)
		if err != nil {
			return diagAndTidy(ctx, client, storage.UUID, err)
		}

		_, err = client.WaitForStorageImportCompletion(ctx, &request.WaitForStorageImportCompletionRequest{
			StorageUUID: storage.UUID,
		})
		if err != nil {
			return diagAndTidy(ctx, client, storage.UUID, err)
		}

		// Imported storage will enter a 'syncing' state for a while. Storage in this
		// state can be used by a server so we will wait for that to allow progress.
		_, err = client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
			UUID:         storage.UUID,
			DesiredState: upcloud.StorageStateSyncing,
		})
		if err != nil {
			return diagAndTidy(ctx, client, storage.UUID, err)
		}
	}

	d.SetId(storage.UUID)

	return diags
}

func isStorageSimpleBackupEnabled(ctx context.Context, service *service.Service, storageID string) (bool, error) {
	details, err := service.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{UUID: storageID})
	if err != nil {
		return false, err
	}
	for _, srvID := range details.ServerUUIDs {
		srv, err := service.GetServerDetails(ctx, &request.GetServerDetailsRequest{UUID: srvID})
		if err != nil {
			return false, err
		}
		if srv.SimpleBackup != "no" {
			return true, nil
		}
	}
	return false, nil
}

func diagAndTidy(ctx context.Context, client *service.Service, storageUUID string, err error) diag.Diagnostics {
	_, waitErr := client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         storageUUID,
		DesiredState: upcloud.StorageStateOnline,
	})
	if waitErr != nil {
		return diag.Errorf("wait for storage after import error: %s", waitErr.Error())
	}

	delErr := client.DeleteStorage(ctx, &request.DeleteStorageRequest{
		UUID: storageUUID,
	})
	if delErr != nil {
		return diag.Errorf("delete storage after import error: %s", delErr.Error())
	}
	return diag.Errorf("storage import error: %s", err.Error())
}
