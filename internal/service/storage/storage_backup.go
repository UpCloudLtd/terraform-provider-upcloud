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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	CreatedAt     types.String `tfsdk:"created_at"`
	storageCommonModel
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
			"encrypt": schema.BoolAttribute{
				MarkdownDescription: encryptDescription,
				Computed:            true,
			},
			"labels":        utils.LabelsAttribute("storage"),
			"system_labels": utils.SystemLabelsAttribute("storage"),
			"size": schema.Int64Attribute{
				MarkdownDescription: sizeDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"tier": schema.StringAttribute{
				MarkdownDescription: tierDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: typeDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "The zone the storage is in, e.g. `de-fra1`.",
				Computed:            true,
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

	var labelsMap map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelsMap, false)...)
	}
	labels := utils.NilAsEmptyList(utils.LabelsMapToSlice(labelsMap))

	if len(labels) > 0 {
		backupDetails, err = r.client.ModifyStorage(ctx, &request.ModifyStorageRequest{
			UUID:   backupDetails.UUID,
			Labels: &labels,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to modify the backup",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
	}

	data.CreatedAt = types.StringValue(backupDetails.Created.Format(time.RFC3339))
	resp.Diagnostics.Append(setCommonValues(ctx, &data.storageCommonModel, &backupDetails.Storage)...)
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

	data.CreatedAt = types.StringValue(backupDetails.Created.Format(time.RFC3339))
	resp.Diagnostics.Append(setCommonValues(ctx, &data.storageCommonModel, &backupDetails.Storage)...)

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

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}
	labelsSlice := utils.NilAsEmptyList(utils.LabelsMapToSlice(labels))

	modifyStorageRequest := request.ModifyStorageRequest{
		Labels: &labelsSlice,
		Title:  data.Title.ValueString(),
		UUID:   state.ID.ValueString(),
	}

	backupDetails, err := r.client.ModifyStorage(ctx, &modifyStorageRequest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update storage backup", utils.ErrorDiagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(setCommonValues(ctx, &data.storageCommonModel, &backupDetails.Storage)...)
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
