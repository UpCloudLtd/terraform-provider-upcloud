package filestorage

import (
	"context"
	"fmt"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &fileStorageResource{}
	_ resource.ResourceWithConfigure   = &fileStorageResource{}
	_ resource.ResourceWithImportState = &fileStorageResource{}

	resourceNameRegexp = regexp.MustCompile(resourceNameRegexpStr)
)

const (
	resourceNameRegexpStr = "^[a-zA-Z0-9_-]+$"
)

func NewFileStorageResource() resource.Resource {
	return &fileStorageResource{}
}

type fileStorageResource struct {
	client *service.Service
}

func (r *fileStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_storage"
}

// Configure adds the provider configured client to the resource.
func (r *fileStorageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type fileStorageModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Size             types.Int64  `tfsdk:"size"`
	Zone             types.String `tfsdk:"zone"`
	ConfiguredStatus types.String `tfsdk:"configured_status"`
	Networks         types.Set    `tfsdk:"network"`
	Labels           types.Map    `tfsdk:"labels"`
}

type networkAttachmentModel struct {
	UUID      types.String `tfsdk:"uuid"`
	Name      types.String `tfsdk:"name"`
	Family    types.String `tfsdk:"family"`
	IPAddress types.String `tfsdk:"ip_address"`
}

var networkAttrTypes = map[string]attr.Type{
	"uuid":       types.StringType,
	"name":       types.StringType,
	"family":     types.StringType,
	"ip_address": types.StringType,
}

func (r *fileStorageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Resource for managing UpCloud file storages. See UpCloud [File Storage](https://upcloud.com/products/file-storage/) product page for more details about the service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the file storage.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the file storage service.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(resourceNameRegexp, fmt.Sprintf("name string that consists only of letters (a–z, A–Z), digits (0–9), underscores (_), or hyphens (-) — with at least one character, and nothing else allowed (no spaces, symbols, or accents): %s", resourceNameRegexp)),
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"size": schema.Int64Attribute{
				Description: "Size of the file storage in GB.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(250),
					int64validator.AtMost(25000),
				},
			},
			"zone": schema.StringAttribute{
				Description: "Zone in which the service will be hosted, e.g. `fi-hel1`. You can list available zones with `upctl zone list`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"configured_status": schema.StringAttribute{
				Description: "The service configured status indicates the service's current intended status. Managed by the customer.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(upcloud.FileStorageConfiguredStatusStarted),
						string(upcloud.FileStorageConfiguredStatusStopped),
					),
				},
			},
			"labels": utils.LabelsAttribute("file storage"),
		},
		Blocks: map[string]schema.Block{
			"network": schema.SetNestedBlock{
				Description: "Network attached to this file storage (currently supports at most one of these blocks).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Description: "UUID of an existing private network to attach.",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Attachment name (unique per this service).",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(resourceNameRegexp, fmt.Sprintf("name string that consists only of letters (a–z, A–Z), digits (0–9), underscores (_), or hyphens (-): %s", resourceNameRegexp)),
								stringvalidator.LengthBetween(1, 64),
							},
						},
						"family": schema.StringAttribute{
							Description: "IP family, e.g. IPv4.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(upcloud.IPAddressFamilyIPv4, upcloud.IPAddressFamilyIPv6),
							},
						},
						"ip_address": schema.StringAttribute{
							Description: "IP address to assign (optional, auto-assign otherwise).",
							Optional:    true,
							Computed:    true,
						},
					},
				},
				Validators: []validator.Set{
					setvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func setFileStorageModel(ctx context.Context, data *fileStorageModel, fileStorage *upcloud.FileStorage) diag.Diagnostics {
	data.ID = types.StringValue(fileStorage.UUID)
	data.Name = types.StringValue(fileStorage.Name)
	data.Size = types.Int64Value(int64(fileStorage.SizeGiB))
	data.Zone = types.StringValue(fileStorage.Zone)
	data.ConfiguredStatus = types.StringValue(string(fileStorage.ConfiguredStatus))
	var diags diag.Diagnostics
	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(fileStorage.Labels))
	if diags.HasError() {
		return diags
	}

	if len(fileStorage.Networks) > 0 {
		networkObjects := make([]attr.Value, len(fileStorage.Networks))
		for i, n := range fileStorage.Networks {
			networkObjects[i], diags = types.ObjectValue(networkAttrTypes, map[string]attr.Value{
				"uuid":       types.StringValue(n.UUID),
				"name":       types.StringValue(n.Name),
				"family":     types.StringValue(n.Family),
				"ip_address": types.StringValue(n.IPAddress),
			})
			if diags.HasError() {
				return diags
			}
		}
		data.Networks, diags = types.SetValue(types.ObjectType{AttrTypes: networkAttrTypes}, networkObjects)
		if diags.HasError() {
			return diags
		}
	}

	return nil
}

func (r *fileStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data fileStorageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Name.IsUnknown() || data.Zone.IsUnknown() || data.Size.IsUnknown() || data.ConfiguredStatus.IsUnknown() {
		resp.Diagnostics.AddError("Invalid plan", "One or more required fields are unknown at apply time.")
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}

	fileStorageRequest := request.CreateFileStorageRequest{
		Name:             data.Name.ValueString(),
		SizeGiB:          int(data.Size.ValueInt64()),
		Zone:             data.Zone.ValueString(),
		ConfiguredStatus: upcloud.FileStorageConfiguredStatus(data.ConfiguredStatus.ValueString()),
		Labels:           utils.LabelsMapToSlice(labels),
	}

	if !data.Networks.IsNull() && !data.Networks.IsUnknown() {
		var nets []networkAttachmentModel
		resp.Diagnostics.Append(data.Networks.ElementsAs(ctx, &nets, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(nets) > 1 {
			resp.Diagnostics.AddError("Invalid number of networks", "Currently only one network attachment is supported.")
			return
		}

		if len(nets) == 1 {
			n := nets[0]
			fileStorageRequest.Networks = []upcloud.FileStorageNetwork{
				{
					UUID:      n.UUID.ValueString(),
					Name:      n.Name.ValueString(),
					Family:    n.Family.ValueString(),
					IPAddress: n.IPAddress.ValueString(),
				},
			}
		}
	}

	fileStorage, err := r.client.CreateFileStorage(ctx, &fileStorageRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create file storage", err.Error())
		return
	}

	waitReq := &request.WaitForFileStorageOperationalStateRequest{
		DesiredState: upcloud.FileStorageOperationalStateRunning,
		UUID:         fileStorage.UUID,
	}

	if data.ConfiguredStatus.ValueString() == string(upcloud.FileStorageConfiguredStatusStopped) {
		waitReq.DesiredState = "stopped"
	}

	fileStorage, err = r.client.WaitForFileStorageOperationalState(ctx, waitReq)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error while waiting for File Storage to be in %s state", waitReq.DesiredState), err.Error())
	}

	resp.Diagnostics.Append(setFileStorageModel(ctx, &data, fileStorage)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *fileStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data fileStorageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	fileStorage, err := r.client.GetFileStorage(ctx, &request.GetFileStorageRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read file storage details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setFileStorageModel(ctx, &data, fileStorage)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *fileStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state fileStorageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Name.IsUnknown() || plan.Zone.IsUnknown() || plan.Size.IsUnknown() || plan.ConfiguredStatus.IsUnknown() {
		resp.Diagnostics.AddError("Invalid plan", "One or more required fields are unknown at apply time.")
		return
	}

	var labels map[string]string
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	}

	uuid := state.ID.ValueString()
	name := plan.Name.ValueString()
	sizeGiB := int(plan.Size.ValueInt64())
	configuredStatus := upcloud.FileStorageConfiguredStatus(plan.ConfiguredStatus.ValueString())
	labelsSlice := utils.LabelsMapToSlice(labels)
	patch := &request.ModifyFileStorageRequest{
		UUID:             uuid,
		Name:             &name,
		SizeGiB:          &sizeGiB,
		ConfiguredStatus: &configuredStatus,
		Labels:           &labelsSlice,
	}

	var planNets, stateNets []networkAttachmentModel
	if !plan.Networks.IsNull() && !plan.Networks.IsUnknown() {
		resp.Diagnostics.Append(plan.Networks.ElementsAs(ctx, &planNets, false)...)
	}
	if !state.Networks.IsNull() && !state.Networks.IsUnknown() {
		resp.Diagnostics.Append(state.Networks.ElementsAs(ctx, &stateNets, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	switch {
	case len(planNets) == 0 && len(stateNets) > 0:
		// Detach all
		patch.Networks = &[]upcloud.FileStorageNetwork{}

	case len(planNets) == 1:
		n := planNets[0]
		var prev networkAttachmentModel
		if len(stateNets) == 1 {
			prev = stateNets[0]
		}

		changed := prev.UUID.ValueString() != n.UUID.ValueString() ||
			prev.Name.ValueString() != n.Name.ValueString() ||
			prev.Family.ValueString() != n.Family.ValueString() ||
			prev.IPAddress.ValueString() != n.IPAddress.ValueString()

		if changed || len(stateNets) == 0 {
			patch.Networks = &[]upcloud.FileStorageNetwork{{
				UUID:      n.UUID.ValueString(),
				Name:      n.Name.ValueString(),
				Family:    n.Family.ValueString(),
				IPAddress: n.IPAddress.ValueString(),
			}}
		}
	}

	fileStorage, err := r.client.ModifyFileStorage(ctx, patch)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update file storage", err.Error())
		return
	}

	waitReq := &request.WaitForFileStorageOperationalStateRequest{
		DesiredState: upcloud.FileStorageOperationalStateRunning,
		UUID:         fileStorage.UUID,
	}

	if plan.ConfiguredStatus.ValueString() == string(upcloud.FileStorageConfiguredStatusStopped) {
		waitReq.DesiredState = "stopped"
	}

	fileStorage, err = r.client.WaitForFileStorageOperationalState(ctx, waitReq)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error while waiting for File Storage to be in %s state", waitReq.DesiredState), err.Error())
	}

	resp.Diagnostics.Append(setFileStorageModel(ctx, &plan, fileStorage)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *fileStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data fileStorageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFileStorage(ctx, &request.DeleteFileStorageRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete file storage", err.Error())
		return
	}

	err = r.client.WaitForFileStorageDeletion(ctx, &request.WaitForFileStorageDeletionRequest{UUID: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("File storage deletion did not complete on time, please check the resource", err.Error())
		return
	}
}

func (r *fileStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
