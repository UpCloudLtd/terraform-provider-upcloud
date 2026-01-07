package filestorage

import (
	"context"
	"fmt"

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
	_ resource.Resource                = &fileStorageShareACLResource{}
	_ resource.ResourceWithConfigure   = &fileStorageShareACLResource{}
	_ resource.ResourceWithImportState = &fileStorageShareACLResource{}
)

func NewFileStorageShareACLResource() resource.Resource {
	return &fileStorageShareACLResource{}
}

type fileStorageShareACLResource struct {
	client *service.Service
}

func (r *fileStorageShareACLResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_storage_share_acl"
}

// Configure adds the provider configured client to the resource.
func (r *fileStorageShareACLResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type fileStorageShareACLModel struct {
	ID          types.String `tfsdk:"id"`
	FileStorage types.String `tfsdk:"file_storage"`
	ShareName   types.String `tfsdk:"share_name"`
	Name        types.String `tfsdk:"name"`
	Target      types.String `tfsdk:"target"`
	Permission  types.String `tfsdk:"permission"`
}

func (r *fileStorageShareACLResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents a File Storage Share ACL entry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the file storage share ACL.",
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
			"share_name": schema.StringAttribute{
				Description: "Name of the share.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(resourceNameRegexp, fmt.Sprintf("share_name string that consists only of letters (a–z, A–Z), digits (0–9), underscores (_), or hyphens (-) — with at least one character, and nothing else allowed (no spaces, symbols, or accents): %s", resourceNameRegexp)),
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"name": schema.StringAttribute{
				Description: "Unique name of the ACL entry (1–64 chars).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(resourceNameRegexp, fmt.Sprintf("name string that consists only of letters (a–z, A–Z), digits (0–9), underscores (_), or hyphens (-) — with at least one character, and nothing else allowed (no spaces, symbols, or accents): %s", resourceNameRegexp)),
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"target": schema.StringAttribute{
				Description: "Target IP/CIDR or '*'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"permission": schema.StringAttribute{
				Description: "Access level: 'ro' or 'rw'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("ro", "rw"),
				},
			},
		},
	}
}

func setFileStorageShareACLModel(_ context.Context, data *fileStorageShareACLModel, fileStorageShareACL *upcloud.FileStorageShareACL) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	var fileStorage, shareName, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &fileStorage, &shareName, &name)
	if err != nil {
		respDiagnostics.AddError(
			"Unable to unmarshal File storage share ACL ID",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	data.FileStorage = types.StringValue(fileStorage)
	data.ShareName = types.StringValue(shareName)
	data.Name = types.StringValue(fileStorageShareACL.Name)
	data.Target = types.StringValue(fileStorageShareACL.Target)
	data.Permission = types.StringValue(string(fileStorageShareACL.Permission))

	return respDiagnostics
}

func (r *fileStorageShareACLResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data fileStorageShareACLModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := &request.CreateFileStorageShareACLRequest{
		ServiceUUID: data.FileStorage.ValueString(),
		ShareName:   data.ShareName.ValueString(),
		FileStorageShareACL: upcloud.FileStorageShareACL{
			Name:       data.Name.ValueString(),
			Target:     data.Target.ValueString(),
			Permission: upcloud.FileStorageShareACLPermission(data.Permission.ValueString()),
		},
	}
	fileStorageShareACL, err := r.client.CreateFileStorageShareACL(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating File Storage share ACL",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.FileStorage.ValueString(), data.ShareName.ValueString(), data.Name.ValueString()))

	resp.Diagnostics.Append(setFileStorageShareACLModel(ctx, &data, fileStorageShareACL)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *fileStorageShareACLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data fileStorageShareACLModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	fileStorageShareACL, err := r.client.GetFileStorageShareACL(ctx, &request.GetFileStorageShareACLRequest{
		ServiceUUID: data.FileStorage.ValueString(),
		ShareName:   data.ShareName.ValueString(),
		ACLName:     data.Name.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read file storage share ACL details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setFileStorageShareACLModel(ctx, &data, fileStorageShareACL)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *fileStorageShareACLResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state fileStorageShareACLModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patchRequest := &request.ModifyFileStorageShareACLRequest{
		ServiceUUID: state.FileStorage.ValueString(),
		ShareName:   state.ShareName.ValueString(),
		ACLName:     state.Name.ValueString(),
		ModifyFileStorageShareACL: request.ModifyFileStorageShareACL{
			Target: upcloud.StringPtr(plan.Target.ValueString()),
			Permission: func() *upcloud.FileStorageShareACLPermission {
				p := upcloud.FileStorageShareACLPermission(plan.Permission.ValueString())
				return &p
			}(),
		},
	}
	fileStorageShareACL, err := r.client.ModifyFileStorageShareACL(ctx, patchRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error updating file storage share ACL", utils.ErrorDiagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(setFileStorageShareACLModel(ctx, &plan, fileStorageShareACL)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *fileStorageShareACLResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data fileStorageShareACLModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFileStorageShareACL(ctx, &request.DeleteFileStorageShareACLRequest{
		ServiceUUID: data.FileStorage.ValueString(),
		ShareName:   data.ShareName.ValueString(),
		ACLName:     data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete file storage share ACL", err.Error())

		return
	}
}

func (r *fileStorageShareACLResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
