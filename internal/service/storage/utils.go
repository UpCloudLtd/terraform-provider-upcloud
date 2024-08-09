package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func ResizeStoragePartitionAndFs(ctx context.Context, client *service.Service, UUID string, deleteBackup bool) diag.Diagnostics {
	var diags diag.Diagnostics

	backup, err := client.ResizeStorageFilesystem(ctx, &request.ResizeStorageFilesystemRequest{
		UUID: UUID,
	})
	if err != nil {
		summary := fmt.Sprintf(
			"Failed to resize partition and filesystem for storage %s",
			UUID,
		)

		diags.AddWarning(
			summary,
			utils.ErrorDiagnosticDetail(err),
		)
	}

	if err == nil && deleteBackup {
		err = client.DeleteStorage(ctx, &request.DeleteStorageRequest{
			UUID: backup.UUID,
		})
		if err != nil {
			summary := fmt.Sprintf(
				"Failed to delete the backup of storage %s after the partition and filesystem resize; you will need to delete the backup manually",
				UUID,
			)

			diags.AddWarning(
				summary,
				utils.ErrorDiagnosticDetail(err),
			)
		}
	}

	return diags
}

func cloneStorage(
	ctx context.Context,
	client *service.Service,
	cloneReq request.CloneStorageRequest,
	modifyReq request.ModifyStorageRequest,
) (*upcloud.StorageDetails, diag.Diagnostics) {
	var diags diag.Diagnostics

	sourceStorage, err := client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         cloneReq.UUID,
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		diags.AddError(
			"Source storage did not reach online state",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}

	if sourceStorage.Size > modifyReq.Size {
		diags.AddError(
			"Unable to clone storage",
			"New storage device should be at least the same size as the source storage.",
		)
		return nil, diags
	}

	storage, err := client.CloneStorage(ctx, &cloneReq)
	if err != nil {
		diags.AddError(
			"Unable to clone storage",
			utils.ErrorDiagnosticDetail(err),
		)
		return storage, diags
	}

	storage, err = client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         storage.UUID,
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		diags.AddError(
			"Cloned storage did not reach online state",
			utils.ErrorDiagnosticDetail(err),
		)
		return storage, diags
	}

	// If the storage specified does not match the cloned storage, modify it so that it does.
	if modifyReq.Size != storage.Size || (modifyReq.Labels != nil && len(*modifyReq.Labels) > 0) || modifyReq.BackupRule != nil {
		modifyReq.UUID = storage.UUID
		storage, err = client.ModifyStorage(ctx, &modifyReq)
		if err != nil {
			diags.AddError(
				"Unable to modify the cloned storage",
				utils.ErrorDiagnosticDetail(err),
			)
			return storage, diags
		}

		storage, err = client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
			UUID:         storage.UUID,
			DesiredState: upcloud.StorageStateOnline,
		})
		if err != nil {
			diags.AddError(
				"Cloned storage did not reach online state",
				utils.ErrorDiagnosticDetail(err),
			)
			return storage, diags
		}
	}

	return storage, diags
}

func templatizeStorage(
	ctx context.Context,
	client *service.Service,
	templatizeReq request.TemplatizeStorageRequest,
	modifyReq request.ModifyStorageRequest,
) (*upcloud.StorageDetails, diag.Diagnostics) {
	var diags diag.Diagnostics

	_, err := client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         templatizeReq.UUID,
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		diags.AddError(
			"Source storage did not reach online state",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}

	storage, err := client.TemplatizeStorage(ctx, &templatizeReq)
	if err != nil {
		diags.AddError(
			"Unable to templatize storage",
			utils.ErrorDiagnosticDetail(err),
		)
		return storage, diags
	}

	storage, err = client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         storage.UUID,
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		diags.AddError(
			"Cloned storage did not reach online state",
			utils.ErrorDiagnosticDetail(err),
		)
		return storage, diags
	}

	// If the storage specified does not match the cloned storage, modify it so that it does.
	if modifyReq.Labels != nil && len(*modifyReq.Labels) > 0 || modifyReq.BackupRule != nil {
		modifyReq.UUID = storage.UUID
		storage, err = client.ModifyStorage(ctx, &modifyReq)
		if err != nil {
			diags.AddError(
				"Unable to modify the cloned storage",
				utils.ErrorDiagnosticDetail(err),
			)
			return storage, diags
		}

		storage, err = client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
			UUID:         storage.UUID,
			DesiredState: upcloud.StorageStateOnline,
		})
		if err != nil {
			diags.AddError(
				"Cloned storage did not reach online state",
				utils.ErrorDiagnosticDetail(err),
			)
			return storage, diags
		}
	}

	return storage, diags
}

func createStorage(
	ctx context.Context,
	client *service.Service,
	createReq request.CreateStorageRequest,
	importReq *request.CreateStorageImportRequest,
) (*upcloud.StorageDetails, diag.Diagnostics) {
	var diags diag.Diagnostics

	storage, err := client.CreateStorage(ctx, &createReq)
	if err != nil {
		diags.AddError(
			"Unable to create storage",
			utils.ErrorDiagnosticDetail(err),
		)
		return storage, diags
	}

	// Wait for storage to enter the 'online' state. For a fresh storage device this is pretty quick.
	storage, err = client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         storage.UUID,
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		diags.AddError(
			"Created storage did not reach online state",
			utils.ErrorDiagnosticDetail(err),
		)
		return storage, diags
	}

	if importReq != nil {
		importReq.StorageUUID = storage.UUID
		_, err := client.CreateStorageImport(ctx, importReq)
		if err != nil {
			return diagAndTidy(ctx, client, storage, err)
		}

		_, err = client.WaitForStorageImportCompletion(ctx, &request.WaitForStorageImportCompletionRequest{
			StorageUUID: storage.UUID,
		})
		if err != nil {
			return diagAndTidy(ctx, client, storage, err)
		}

		// Imported storage will enter a 'syncing' state for a while. Storage in this state can be used by a server so we will wait for that to allow progress.
		_, err = client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
			UUID:         storage.UUID,
			DesiredState: upcloud.StorageStateSyncing,
		})
		if err != nil {
			return diagAndTidy(ctx, client, storage, err)
		}
	}

	return storage, diags
}

func diagAndTidy(ctx context.Context, client *service.Service, storage *upcloud.StorageDetails, err error) (*upcloud.StorageDetails, diag.Diagnostics) {
	var diags diag.Diagnostics
	diags.AddError(
		"Storage import failed",
		utils.ErrorDiagnosticDetail(err),
	)

	_, waitErr := client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         storage.UUID,
		DesiredState: upcloud.StorageStateOnline,
	})
	if waitErr != nil {
		diags.AddError(
			"Created storage did not reach online state after failed import",
			utils.ErrorDiagnosticDetail(waitErr),
		)
		return storage, diags
	}

	delErr := client.DeleteStorage(ctx, &request.DeleteStorageRequest{
		UUID: storage.UUID,
	})
	if delErr != nil {
		diags.AddError(
			"Unable to delete storage after failed import",
			utils.ErrorDiagnosticDetail(delErr),
		)
		return storage, diags
	}

	return storage, diags
}

func checkHash(sourceHash string, importDetails *upcloud.StorageImportDetails) error {
	if sourceHash == "" {
		return nil
	}

	hash := strings.SplitN(sourceHash, " ", 2)[0]

	if hash != importDetails.SHA256Sum {
		return fmt.Errorf("imported storage's SHA256 sum does not match the source_hash: expected %s, got %s", hash, importDetails.SHA256Sum)
	}

	return nil
}
