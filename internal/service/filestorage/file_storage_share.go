package filestorage

import (
	"context"
	"fmt"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

	sharePathRegexp = regexp.MustCompile(sharePathRegexpStr)
)

const (
	sharePathRegexpStr = "^/[a-z0-9/_-]*$"
)

func NewFileStorageShareResource() resource.Resource {
	return &fileStorageShareResource{}
}

type fileStorageShareResource struct {
	client *service.Service
}

func (r *fileStorageShareResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_storage_share"
}

// Configure adds the provider configured client to the resource.
func (r *fileStorageShareResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type fileStorageShareModel struct {
	ID          types.String `tfsdk:"id"`
	FileStorage types.String `tfsdk:"file_storage"`
	Name        types.String `tfsdk:"name"`
	Path        types.String `tfsdk:"path"`
}

func (r *fileStorageShareResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents a File Storage service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the file storage share.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file_storage": schema.StringAttribute{
				Description: "UUID of the file storage service.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Unique name of the share (1–64 chars).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(resourceNameRegexp, fmt.Sprintf("name string that consists only of letters (a–z, A–Z), digits (0–9), underscores (_), or hyphens (-) — with at least one character, and nothing else allowed (no spaces, symbols, or accents): %s", resourceNameRegexp)),
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"path": schema.StringAttribute{
				Description: "Absolute path exported by the share (e.g. `/public`).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(sharePathRegexp, fmt.Sprintf("path string that consists only of lower case letters (a–z), digits (0–9), underscores (_), or hyphens (-) — with at least one character, and nothing else allowed (no spaces, symbols, or accents): %s", resourceNameRegexp)),
					stringvalidator.LengthBetween(1, 512),
				},
			},
		},
	}
}

func setFileStorageShareModel(_ context.Context, data *fileStorageShareModel, fileStorageShare *upcloud.FileStorageShare) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	var fileStorage, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &fileStorage, &name)
	if err != nil {
		respDiagnostics.AddError(
			"Unable to unmarshal File storage share name",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	data.FileStorage = types.StringValue(fileStorage)
	data.Path = types.StringValue(fileStorageShare.Path)
	data.Name = types.StringValue(fileStorageShare.Name)

	return respDiagnostics
}

func (r *fileStorageShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data fileStorageShareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := &request.CreateFileStorageShareRequest{
		ServiceUUID: data.FileStorage.ValueString(),
		Name:        data.Name.ValueString(),
		Path:        data.Path.ValueString(),
		ACL:         []upcloud.FileStorageShareACL{},
	}
	fileStorageShare, err := r.client.CreateFileStorageShare(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating File Storage share",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.FileStorage.ValueString(), data.Name.ValueString()))

	resp.Diagnostics.Append(setFileStorageShareModel(ctx, &data, fileStorageShare)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *fileStorageShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data fileStorageShareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	fileStorageShare, err := r.client.GetFileStorageShare(ctx, &request.GetFileStorageShareRequest{
		ServiceUUID: data.FileStorage.ValueString(),
		ShareName:   data.Name.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read file storage share details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setFileStorageShareModel(ctx, &data, fileStorageShare)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *fileStorageShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state fileStorageShareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fileStorage := state.FileStorage.ValueString()
	stateName := state.Name.ValueString()
	planName := plan.Name.ValueString()

	patchRequest := &request.ModifyFileStorageShareRequest{
		ServiceUUID: fileStorage,
		ShareName:   stateName,
		ModifyFileStorageShare: request.ModifyFileStorageShare{
			Name: upcloud.StringPtr(planName),
		},
	}
	fileStorageShare, err := r.client.ModifyFileStorageShare(ctx, patchRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error updating file storage share", utils.ErrorDiagnosticDetail(err))

		return
	}

	resp.Diagnostics.Append(setFileStorageShareModel(ctx, &plan, fileStorageShare)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *fileStorageShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data fileStorageShareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFileStorageShare(ctx, &request.DeleteFileStorageShareRequest{
		ServiceUUID: data.FileStorage.ValueString(),
		ShareName:   data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete file storage share", err.Error())

		return
	}
}

func (r *fileStorageShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
