package managedobjectstorage

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &managedObjectStorageUserResource{}
	_ resource.ResourceWithConfigure   = &managedObjectStorageUserResource{}
	_ resource.ResourceWithImportState = &managedObjectStorageUserResource{}
)

func NewUserResource() resource.Resource {
	return &managedObjectStorageUserResource{}
}

type managedObjectStorageUserResource struct {
	client *service.Service
}

func (r *managedObjectStorageUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage_user"
}

// Configure adds the provider configured client to the resource.
func (r *managedObjectStorageUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type userModel struct {
	ARN         types.String `tfsdk:"arn"`
	CreatedAt   types.String `tfsdk:"created_at"`
	ID          types.String `tfsdk:"id"`
	ServiceUUID types.String `tfsdk:"service_uuid"`
	Username    types.String `tfsdk:"username"`
}

func (r *managedObjectStorageUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an UpCloud Managed Object Storage user. No relation to UpCloud API accounts.",
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Description: "User ARN.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Creation time.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Description: "ID of the user. ID is in {object storage UUID}/{username} format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Custom usernames for accessing the object storage. No relation to UpCloud API accounts. See `upcloud_managed_object_storage_user_access_key` for managing access keys and `upcloud_managed_object_storage_user_policy` for managing policies.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_uuid": schema.StringAttribute{
				Description: "Managed Object Storage service UUID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func setUserValues(_ context.Context, data *userModel, user *upcloud.ManagedObjectStorageUser) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	data.ARN = types.StringValue(user.ARN)
	data.CreatedAt = types.StringValue(user.CreatedAt.String())
	data.Username = types.StringValue(user.Username)

	return respDiagnostics
}

func (r *managedObjectStorageUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.Username.ValueString()))

	apiReq := &request.CreateManagedObjectStorageUserRequest{
		Username:    data.Username.ValueString(),
		ServiceUUID: data.ServiceUUID.ValueString(),
	}

	user, err := r.client.CreateManagedObjectStorageUser(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage user",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setUserValues(ctx, &data, user)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	var serviceUUID, username string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &username)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ServiceUUID = types.StringValue(serviceUUID)

	user, err := r.client.GetManagedObjectStorageUser(ctx, &request.GetManagedObjectStorageUserRequest{
		Username:    username,
		ServiceUUID: serviceUUID,
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read managed object storage user details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setUserValues(ctx, &data, user)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All configurable attributes require replace, so Update method is only required to satisfy the interface.
}

func (r *managedObjectStorageUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var serviceUUID, username string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &username)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteManagedObjectStorageUser(ctx, &request.DeleteManagedObjectStorageUserRequest{
		ServiceUUID: serviceUUID,
		Username:    username,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage user",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *managedObjectStorageUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
