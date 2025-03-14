package storage

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

var (
	_ resource.Resource                = &storageBackupResource{}
	_ resource.ResourceWithConfigure   = &storageBackupResource{}
	_ resource.ResourceWithImportState = &storageBackupResource{}
)

type storageBackupResource struct {
	client *service.Service
}

func NewStorageBackupResource() resource.Resource {
	return &storageBackupResource{}
}

func (r *storageBackupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_backup"
}

// Configure adds the provider configured client to the resource.
func (r *storageBackupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type storageBackupModel struct {
	SourceStorage types.String `tfsdk:"source_storage"`
	Title         types.String `tfsdk:"title"`
	ID            types.String `tfsdk:"id"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

func (r *storageBackupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates an on-demand storage backup in UpCloud. To force a backup, change the backup title to a unique name",
		Attributes: map[string]schema.Attribute{
			"source_storage": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the storage to back up.",
			},
			"title": schema.StringAttribute{
				Required:    true,
				Description: "Title of the backup.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the created backup.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the backup creation.",
			},
		},
	}
}

func (r *storageBackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data storageBackupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a backup
	_, err := r.client.CreateBackup(ctx, &request.CreateBackupRequest{
		UUID:  data.SourceStorage.ValueString(),
		Title: data.Title.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Failed to create storage backup", utils.ErrorDiagnosticDetail(err))
		return
	}

	/*
		// Format storage details as JSON for better readability
		storageDetailsJSON, err := json.MarshalIndent(storageDetails, "", "  ")
		if err != nil {
			tflog.Error(ctx, "Failed to format storage details", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			tflog.Info(ctx, "Retrieved storage details:\n"+string(storageDetailsJSON))
		}

		tflog.Info(ctx, "--------------------------------------------------- 3")
	*/

	// Wait for backup to finish
	_, err = r.client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         data.SourceStorage.ValueString(),
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Storage did not reach online state after backup process",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	/*
		// Check if backup already exists
		backup, err := r.client.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{
			UUID: data.SourceStorage.ValueString(),
		})

		if err != nil {
			resp.Diagnostics.AddError("Storage not found", utils.ErrorDiagnosticDetail(err))
			return
		}

		// Format storage details as JSON for better readability
		backupJSON, err := json.MarshalIndent(backup, "", "  ")
		if err != nil {
			tflog.Error(ctx, "Failed to format storage details", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			tflog.Info(ctx, "Retrieved storage details:\n"+string(backupJSON))
		}
	*/
}

func (r *storageBackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data storageBackupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.SourceStorage.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	// Get storage details
	storageDetails, err := r.client.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{
		UUID: data.SourceStorage.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read storage details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	// Check if backup exists
	if !slices.Contains(storageDetails.BackupUUIDs, data.ID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}

	// Get the Storage Backup details - IT DOES NOT EXIST
	backup, err := r.client.GetStorageBackupDetails(ctx, &request.GetStorageBackupDetailsRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read backup storage details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	data.CreatedAt = types.StringValue(backup.Created)
	data.Title = types.StringValue(backup.Title)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *storageBackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// To be implemented
}

func (r *storageBackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// To be implemented
}

func (r *storageBackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// To be implemented
}
