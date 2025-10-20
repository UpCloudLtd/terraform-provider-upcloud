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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	Network          types.Object `tfsdk:"network"`
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
		Description: "Resource for managing UpCloud file storages.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the file storage.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the file storage.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(resourceNameRegexp, fmt.Sprintf("name string that consists only of letters (a–z, A–Z), digits (0–9), underscores (_), or hyphens (-) — with at least one character, and nothing else allowed (no spaces, symbols, or accents): %s", resourceNameRegexp)),
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Size of the file storage in GB.",
				Required:            true,
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
				MarkdownDescription: "The service configured status indicates the service's current intended status. Managed by the customer.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(upcloud.FileStorageConfiguredStatusStarted),
						string(upcloud.FileStorageConfiguredStatusStopped),
					),
				},
			},
			"network": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"uuid": schema.StringAttribute{
						Description: "UUID of an existing private network to attach",
						Required:    true,
					},
					"name": schema.StringAttribute{
						Description: "Attachment name (unique per this service)",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(resourceNameRegexp, fmt.Sprintf("name string that consists only of letters (a–z, A–Z), digits (0–9), underscores (_), or hyphens (-) — with at least one character, and nothing else allowed (no spaces, symbols, or accents): %s", resourceNameRegexp)),
							stringvalidator.LengthBetween(1, 64),
						},
					},
					"family": schema.StringAttribute{
						Description: "IP family, e.g. IPv4",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(upcloud.IPAddressFamilyIPv4, upcloud.IPAddressFamilyIPv6),
						},
					},
					"ip_address": schema.StringAttribute{
						Description: "IP address to assign (optional, auto-assign otherwise)",
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	}
}

func setFileStorageModel(_ context.Context, data *fileStorageModel, fileStorage *upcloud.FileStorage) diag.Diagnostics {
	data.ID = types.StringValue(fileStorage.UUID)
	data.Name = types.StringValue(fileStorage.Name)
	data.Size = types.Int64Value(int64(fileStorage.SizeGiB))
	data.Zone = types.StringValue(fileStorage.Zone)
	data.ConfiguredStatus = types.StringValue(string(fileStorage.ConfiguredStatus))
	if len(fileStorage.Networks) > 0 {
		var diags diag.Diagnostics
		data.Network, diags = types.ObjectValue(networkAttrTypes, map[string]attr.Value{
			"uuid":       types.StringValue(fileStorage.Networks[0].UUID),
			"name":       types.StringValue(fileStorage.Networks[0].Name),
			"family":     types.StringValue(fileStorage.Networks[0].Family),
			"ip_address": types.StringValue(fileStorage.Networks[0].IPAddress),
		})
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

	fileStorageRequest := request.CreateFileStorageRequest{
		Name:             data.Name.ValueString(),
		SizeGiB:          int(data.Size.ValueInt64()),
		Zone:             data.Zone.ValueString(),
		ConfiguredStatus: upcloud.FileStorageConfiguredStatus(data.ConfiguredStatus.ValueString()),
	}

	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		var net networkAttachmentModel
		resp.Diagnostics.Append(data.Network.As(ctx, &net, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		if net.UUID.IsUnknown() || net.Name.IsUnknown() || net.Family.IsUnknown() {
			resp.Diagnostics.AddError(
				"Invalid network block",
				"One or more required fields in 'network' are not known at apply time.",
			)
			return
		}

		fileStorageRequest.Networks = []upcloud.FileStorageNetwork{
			{
				UUID:      net.UUID.ValueString(),
				Name:      net.Name.ValueString(),
				Family:    net.Family.ValueString(),
				IPAddress: net.IPAddress.ValueString(),
			},
		}
	}

	fileStorage, err := r.client.CreateFileStorage(ctx, &fileStorageRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create file storage", err.Error())
		return
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

	uuid := state.ID.ValueString()
	name := plan.Name.ValueString()
	sizeGiB := int(plan.Size.ValueInt64())
	configuredStatus := upcloud.FileStorageConfiguredStatus(plan.ConfiguredStatus.ValueString())
	patch := &request.ModifyFileStorageRequest{
		UUID:             uuid,
		Name:             &name,
		SizeGiB:          &sizeGiB,
		ConfiguredStatus: &configuredStatus,
	}

	if plan.Network.IsNull() && !state.Network.IsNull() {
		patch.Networks = &[]upcloud.FileStorageNetwork{}
	} else if !plan.Network.IsNull() && !plan.Network.IsUnknown() {
		var net networkAttachmentModel
		resp.Diagnostics.Append(plan.Network.As(ctx, &net, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Compare against previous state to see if something changed
		var prev networkAttachmentModel
		if !state.Network.IsNull() && !state.Network.IsUnknown() {
			resp.Diagnostics.Append(state.Network.As(ctx, &prev, basetypes.ObjectAsOptions{})...)
			if resp.Diagnostics.HasError() {
				return
			}
		}

		changed := prev.UUID.ValueString() != net.UUID.ValueString() ||
			prev.Name.ValueString() != net.Name.ValueString() ||
			prev.Family.ValueString() != net.Family.ValueString() ||
			prev.IPAddress.ValueString() != net.IPAddress.ValueString()

		if changed || state.Network.IsNull() {
			patch.Networks = &[]upcloud.FileStorageNetwork{
				{
					UUID:      net.UUID.ValueString(),
					Name:      net.Name.ValueString(),
					Family:    net.Family.ValueString(),
					IPAddress: net.IPAddress.ValueString(),
				},
			}
		}
	}

	fileStorage, err := r.client.ModifyFileStorage(ctx, patch)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update file storage", err.Error())
		return
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
}

func (r *fileStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
