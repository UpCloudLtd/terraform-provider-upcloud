package storage

import (
	"context"
	"strings"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &storageResource{}
	_ resource.ResourceWithConfigure   = &storageResource{}
	_ resource.ResourceWithImportState = &storageResource{}
)

func NewStorageResource() resource.Resource {
	return &storageResource{}
}

type storageResource struct {
	client *service.Service
}

func (r *storageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage"
}

// Configure adds the provider configured client to the resource.
func (r *storageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type storageModel struct {
	storageCommonModel

	BackupRule             types.List `tfsdk:"backup_rule"`
	Clone                  types.Set  `tfsdk:"clone"`
	DeleteAutoresizeBackup types.Bool `tfsdk:"delete_autoresize_backup"`
	FilesystemAutoresize   types.Bool `tfsdk:"filesystem_autoresize"`
	Import                 types.Set  `tfsdk:"import"`
}

type cloneModel struct {
	ID types.String `tfsdk:"id"`
}

type importModel struct {
	Source         types.String `tfsdk:"source"`
	SourceLocation types.String `tfsdk:"source_location"`
	SourceHash     types.String `tfsdk:"source_hash"`
	SHA256sum      types.String `tfsdk:"sha256sum"`
	WrittenBytes   types.Int64  `tfsdk:"written_bytes"`
}

func (r *storageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages UpCloud [Block Storage](https://upcloud.com/products/block-storage) devices.",
		Attributes: map[string]schema.Attribute{
			"delete_autoresize_backup": schema.BoolAttribute{
				Description: "If set to true, the backup taken before the partition and filesystem resize attempt will be deleted immediately after success.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"encrypt": schema.BoolAttribute{
				MarkdownDescription: "Sets if the storage is encrypted at rest.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"filesystem_autoresize": schema.BoolAttribute{
				MarkdownDescription: `If set to true, provider will attempt to resize partition and filesystem when the size of the storage changes. Please note that before the resize attempt is made, backup of the storage will be taken. If the resize attempt fails, the backup will be used to restore the storage and then deleted. If the resize attempt succeeds, backup will be kept (unless ` + "`" + `delete_autoresize_backup` + "`" + ` option is set to true).
				Taking and keeping backups incure costs.`,
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the storage.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"labels":        utils.LabelsAttribute("storage"),
			"system_labels": utils.SystemLabelsAttribute("storage"),
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the storage in gigabytes.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 4096),
				},
			},
			"tier": schema.StringAttribute{
				MarkdownDescription: "The tier of the storage.",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						upcloud.StorageTierMaxIOPS,
						upcloud.StorageTierStandard,
						upcloud.StorageTierHDD,
					),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "A short, informative description.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the storage.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.StringAttribute{
				Description: "The zone the storage is in, e.g. `de-fra1`. You can list available zones with `upctl zone list`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"clone": schema.SetNestedBlock{
				Description: "Block defining another storage/template to clone to storage.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the storage/template to clone.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
				Validators: []validator.Set{
					setvalidator.ConflictsWith(path.MatchRoot("import")),
					setvalidator.SizeBetween(0, 1),
				},
			},
			"import": schema.SetNestedBlock{
				Description: "Block defining external data to import to storage",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							MarkdownDescription: "The mode of the import task. One of `http_import` or `direct_upload`.",
							Required:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(
									upcloud.StorageImportSourceDirectUpload,
									upcloud.StorageImportSourceHTTPImport,
								),
							},
						},
						"source_location": schema.StringAttribute{
							MarkdownDescription: "The location of the file to import. For `http_import` an accessible URL for `direct_upload` a local file.",
							Required:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"source_hash": schema.StringAttribute{
							MarkdownDescription: "SHA256 hash of the source content. This hash is used to verify the integrity of the imported data by comparing it to `sha256sum` after the import has completed. Possible filename is automatically removed from the hash before comparison.",
							Optional:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"sha256sum": schema.StringAttribute{
							Description: "sha256 sum of the imported data",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"written_bytes": schema.Int64Attribute{
							Description: "Number of bytes imported",
							Computed:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
					},
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.ConflictsWith(path.MatchRoot("clone")),
					setvalidator.SizeBetween(0, 1),
				},
			},
			"backup_rule": BackupRuleBlock(),
		},
	}
}

func setStorageValues(ctx context.Context, data *storageModel, storage *upcloud.StorageDetails) diag.Diagnostics {
	respDiagnostics := setCommonValues(ctx, &data.storageCommonModel, storage)

	if !data.BackupRule.IsNull() && storage.BackupRule != nil {
		backupRule := []BackupRuleModel{{
			Interval:  types.StringValue(storage.BackupRule.Interval),
			Time:      types.StringValue(storage.BackupRule.Time),
			Retention: types.Int64Value(int64(storage.BackupRule.Retention)),
		}}

		var diags diag.Diagnostics
		data.BackupRule, diags = types.ListValueFrom(ctx, data.BackupRule.ElementType(ctx), backupRule)
		respDiagnostics.Append(diags...)
	}

	if data.FilesystemAutoresize.IsNull() {
		data.FilesystemAutoresize = types.BoolValue(false)
	}

	if data.DeleteAutoresizeBackup.IsNull() {
		data.DeleteAutoresizeBackup = types.BoolValue(false)
	}

	return respDiagnostics
}

func buildBackupRule(ctx context.Context, dataBackupRule types.List) (*upcloud.BackupRule, diag.Diagnostics) {
	backupRule, diags := utils.GetFirstItem[BackupRuleModel](ctx, dataBackupRule)
	if backupRule == nil {
		return &upcloud.BackupRule{}, diags
	}

	return &upcloud.BackupRule{
		Interval:  backupRule.Interval.ValueString(),
		Time:      backupRule.Time.ValueString(),
		Retention: int(backupRule.Retention.ValueInt64()),
	}, diags
}

func upcloudBoolean(b types.Bool) upcloud.Boolean {
	if b.IsNull() || b.IsUnknown() {
		return upcloud.Empty
	}
	return upcloud.FromBool(b.ValueBool())
}

func (r *storageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data storageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labelsMap map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelsMap, false)...)
	}
	labels := utils.NilAsEmptyList(utils.LabelsMapToSlice(labelsMap))

	var storage *upcloud.StorageDetails
	if !data.Clone.IsNull() {
		var planClone []cloneModel
		resp.Diagnostics.Append(data.Clone.ElementsAs(ctx, &planClone, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		backupRule, diags := buildBackupRule(ctx, data.BackupRule)
		resp.Diagnostics.Append(diags...)

		storage, diags = cloneStorage(ctx, r.client, request.CloneStorageRequest{
			Encrypted: upcloudBoolean(data.Encrypt),
			Tier:      data.Tier.ValueString(),
			Title:     data.Title.ValueString(),
			UUID:      planClone[0].ID.ValueString(),
			Zone:      data.Zone.ValueString(),
		}, request.ModifyStorageRequest{
			BackupRule: backupRule,
			Labels:     &labels,
			Size:       int(data.Size.ValueInt64()),
		})
		resp.Diagnostics.Append(diags...)
	} else {
		var importReq *request.CreateStorageImportRequest
		if !data.Import.IsNull() {
			var planImport []importModel
			resp.Diagnostics.Append(data.Import.ElementsAs(ctx, &planImport, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if len(planImport) == 1 {
				importReq = &request.CreateStorageImportRequest{
					Source:         planImport[0].Source.ValueString(),
					SourceLocation: planImport[0].SourceLocation.ValueString(),
				}
			}
		}

		backupRule, diags := buildBackupRule(ctx, data.BackupRule)
		resp.Diagnostics.Append(diags...)
		storage, diags = createStorage(ctx, r.client, request.CreateStorageRequest{
			BackupRule: backupRule,
			Encrypted:  upcloudBoolean(data.Encrypt),
			Title:      data.Title.ValueString(),
			Size:       int(data.Size.ValueInt64()),
			Tier:       data.Tier.ValueString(),
			Zone:       data.Zone.ValueString(),
			Labels:     labels,
		}, importReq)
		resp.Diagnostics.Append(diags...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setStorageValues(ctx, &data, storage)...)

	if !data.Import.IsNull() {
		importDetails, err := r.client.GetStorageImportDetails(ctx, &request.GetStorageImportDetailsRequest{
			UUID: storage.UUID,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read import details",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}

		planImport, diags := utils.GetFirstItem[importModel](ctx, data.Import)
		resp.Diagnostics.Append(diags...)
		planImport.WrittenBytes = types.Int64Value(int64(importDetails.WrittenBytes))
		planImport.SHA256sum = types.StringValue(importDetails.SHA256Sum)

		err = checkHash(planImport.SourceHash.ValueString(), importDetails)
		if err != nil {
			resp.Diagnostics.AddError(
				"Storage import failed",
				utils.ErrorDiagnosticDetail(err),
			)
		}

		data.Import, diags = types.SetValueFrom(ctx, data.Import.ElementType(ctx), []importModel{*planImport})
		resp.Diagnostics.Append(diags...)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *storageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data storageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	storage, err := r.client.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{
		UUID: data.ID.ValueString(),
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

	resp.Diagnostics.Append(setStorageValues(ctx, &data, storage)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *storageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state storageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	storage, err := r.client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         data.ID.ValueString(),
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Storage did not reach online state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}
	labelsSlice := utils.NilAsEmptyList(utils.LabelsMapToSlice(labels))

	var backupRule *upcloud.BackupRule
	// Do not remove backup rule that has been created outside of Terraform
	if !data.BackupRule.IsNull() || !state.BackupRule.IsNull() {
		var diags diag.Diagnostics
		backupRule, diags = buildBackupRule(ctx, data.BackupRule)
		resp.Diagnostics.Append(diags...)
	}

	apiReq := request.ModifyStorageRequest{
		BackupRule: backupRule,
		Labels:     &labelsSlice,
		Size:       int(data.Size.ValueInt64()),
		Title:      data.Title.ValueString(),
		UUID:       data.ID.ValueString(),
	}

	// Attached server must be stopped before resizing the storage
	sizeChanged := !data.Size.Equal(state.Size)
	stopServer := len(storage.ServerUUIDs) > 0 && sizeChanged
	if stopServer {
		err := utils.VerifyServerStopped(ctx, request.StopServerRequest{UUID: storage.ServerUUIDs[0]}, r.client)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to stop attached server while resizing storage",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
	}

	storage, err = r.client.ModifyStorage(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify storage",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if sizeChanged && data.FilesystemAutoresize.ValueBool() {
		resp.Diagnostics.Append(ResizeStoragePartitionAndFs(ctx, r.client, storage.UUID, data.DeleteAutoresizeBackup.ValueBool())...)
	}

	// Restart the attached server if it was stopped
	if stopServer {
		err := utils.VerifyServerStarted(ctx, request.StartServerRequest{UUID: storage.ServerUUIDs[0]}, r.client)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to restart attached server after resizing storage",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
	}

	resp.Diagnostics.Append(setStorageValues(ctx, &data, storage)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *storageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data storageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	id := data.ID.ValueString()

	// Wait for storage to enter 'online' state as storage devices can only be deleted in this state.
	storage, err := r.client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         id,
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Storage did not reach online state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if len(storage.ServerUUIDs) > 0 {
		serverUUID := storage.ServerUUIDs[0]
		// Get server details for retrieving the address that is to be used when detaching the storage
		serverDetails, err := r.client.GetServerDetails(ctx, &request.GetServerDetailsRequest{
			UUID: serverUUID,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read server details while deleting attached storage",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}

		if storageDevice := serverDetails.StorageDevice(id); storageDevice != nil {
			// ide devices can only be detached from stopped servers
			if strings.HasPrefix(storageDevice.Address, "ide") {
				err = utils.VerifyServerStopped(ctx, request.StopServerRequest{UUID: serverUUID}, r.client)
				if err != nil {
					resp.Diagnostics.AddError(
						"Unable to stop attached server while resizing storage",
						utils.ErrorDiagnosticDetail(err),
					)
					return
				}
			}

			_, err = utils.WithRetry(func() (interface{}, error) {
				return r.client.DetachStorage(ctx, &request.DetachStorageRequest{ServerUUID: serverUUID, Address: storageDevice.Address})
			}, 20, time.Second*3)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to detach storage",
					utils.ErrorDiagnosticDetail(err),
				)
				return
			}

			if strings.HasPrefix(storageDevice.Address, "ide") && serverDetails.State != upcloud.ServerStateStopped {
				// No need to pass host explicitly here, as the server will be started on old host by default (for private clouds)
				if err = utils.VerifyServerStarted(ctx, request.StartServerRequest{UUID: serverUUID}, r.client); err != nil {
					resp.Diagnostics.AddError(
						"Unable to restart attached server after detaching storage",
						utils.ErrorDiagnosticDetail(err),
					)
					return
				}
			}
		}
	}

	deleteStorageRequest := &request.DeleteStorageRequest{
		UUID: id,
	}
	err = r.client.DeleteStorage(ctx, deleteStorageRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete storage",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *storageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
