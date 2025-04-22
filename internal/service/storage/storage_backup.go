package storage

import (
	"context"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (r *storageBackupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an on-demand storage backup.",
		Attributes: map[string]schema.Attribute{
			"source_storage": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the storage to back up.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				Required:    true,
				Description: "Title of the backup.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the created backup.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the backup creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
	backupDetails, err := r.client.CreateBackup(ctx, &request.CreateBackupRequest{
		UUID:  data.SourceStorage.ValueString(),
		Title: data.Title.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create storage backup", utils.ErrorDiagnosticDetail(err))
		return
	}

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

	data.CreatedAt = types.StringValue(backupDetails.Created.Format(time.RFC3339))
	data.ID = types.StringValue(backupDetails.UUID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *storageBackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data storageBackupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	// Get backup details
	backupDetails, err := r.client.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read backup details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	data.ID = types.StringValue(backupDetails.UUID)
	data.CreatedAt = types.StringValue(backupDetails.Created.Format(time.RFC3339))
	data.Title = types.StringValue(backupDetails.Title)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *storageBackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read current configuration
	var data storageBackupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the current state
	var state storageBackupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequired := false
	modifyStorageRequest := request.ModifyStorageRequest{
		UUID: state.ID.ValueString(),
	}

	if data.Title.ValueString() != state.Title.ValueString() {
		modifyStorageRequest.Title = data.Title.ValueString()
		updateRequired = true
	}

	if updateRequired {
		_, err := r.client.ModifyStorage(ctx, &modifyStorageRequest)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update storage backup", utils.ErrorDiagnosticDetail(err))
			return
		}
	}

	data.ID = state.ID
	data.CreatedAt = state.CreatedAt
	data.SourceStorage = state.SourceStorage

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *storageBackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageBackupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	backupID := state.ID.ValueString()

	// Wait for backup to enter 'online' state as storage devices can only be deleted in this state.
	_, err := r.client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         backupID,
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Storage did not reach online state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	// Do we want to delete the snapshot from the system or only make TF think it is deleted?
	deleteStorageRequest := &request.DeleteStorageRequest{
		UUID: backupID,
	}
	err = r.client.DeleteStorage(ctx, deleteStorageRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete backup",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	resp.State.RemoveResource(ctx)
}

func (r *storageBackupResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
