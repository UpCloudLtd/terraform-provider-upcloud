package managedobjectstorage

import (
	"context"

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
	_ resource.Resource                = &managedObjectStorageUserAccessKeyResource{}
	_ resource.ResourceWithConfigure   = &managedObjectStorageUserAccessKeyResource{}
	_ resource.ResourceWithImportState = &managedObjectStorageUserAccessKeyResource{}
)

func NewManagedObjectStorageUserAccessKeyResource() resource.Resource {
	return &managedObjectStorageUserAccessKeyResource{}
}

type managedObjectStorageUserAccessKeyResource struct {
	client *service.Service
}

func (r *managedObjectStorageUserAccessKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage_user_access_key"
}

// Configure adds the provider configured client to the resource.
func (r *managedObjectStorageUserAccessKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type userAccessKeyModel struct {
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	CreatedAt       types.String `tfsdk:"created_at"`
	ID              types.String `tfsdk:"id"`
	LastUsedAt      types.String `tfsdk:"last_used_at"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	ServiceUUID     types.String `tfsdk:"service_uuid"`
	Status          types.String `tfsdk:"status"`
	Username        types.String `tfsdk:"username"`
}

func (r *managedObjectStorageUserAccessKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an UpCloud Managed Object Storage user access key.",
		Attributes: map[string]schema.Attribute{
			"access_key_id": schema.StringAttribute{
				Description: "Access key ID.",
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
				Description: "ID of the user. ID is in {object storage UUID}/{username}/{access key id} format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_used_at": schema.StringAttribute{
				Description: "Last used.",
				Computed:    true,
			},
			"secret_access_key": schema.StringAttribute{
				Description: "Secret access key.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_uuid": schema.StringAttribute{
				Description: "Managed Object Storage service UUID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Status of the key. Valid values: `Active`|`Inactive`",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(upcloud.ManagedObjectStorageUserAccessKeyStatusActive),
						string(upcloud.ManagedObjectStorageUserAccessKeyStatusInactive),
					),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func setUserAccessKeyValues(_ context.Context, data *userAccessKeyModel, accessKey *upcloud.ManagedObjectStorageUserAccessKey) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	data.AccessKeyID = types.StringValue(accessKey.AccessKeyID)
	data.CreatedAt = types.StringValue(accessKey.CreatedAt.String())
	data.LastUsedAt = types.StringValue(accessKey.LastUsedAt.String())

	if accessKey.SecretAccessKey != nil {
		data.SecretAccessKey = types.StringValue(*accessKey.SecretAccessKey)
	}

	return respDiagnostics
}

func (r *managedObjectStorageUserAccessKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data userAccessKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := &request.CreateManagedObjectStorageUserAccessKeyRequest{
		Username:    data.Username.ValueString(),
		ServiceUUID: data.ServiceUUID.ValueString(),
	}

	accessKey, err := r.client.CreateManagedObjectStorageUserAccessKey(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage user access key",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.Username.ValueString(), accessKey.AccessKeyID))

	resp.Diagnostics.Append(setUserAccessKeyValues(ctx, &data, accessKey)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageUserAccessKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data userAccessKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	var serviceUUID, username, accessKeyID string
	err := utils.UnmarshalID(data.ID.ValueString(), &serviceUUID, &username, &accessKeyID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal managed object storage user access key ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ServiceUUID = types.StringValue(serviceUUID)
	data.Username = types.StringValue(username)
	data.AccessKeyID = types.StringValue(accessKeyID)

	user, err := r.client.GetManagedObjectStorageUserAccessKey(ctx, &request.GetManagedObjectStorageUserAccessKeyRequest{
		Username:    username,
		ServiceUUID: serviceUUID,
		AccessKeyID: accessKeyID,
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read managed object storage user access key details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setUserAccessKeyValues(ctx, &data, user)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageUserAccessKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data userAccessKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	accessKey, err := r.client.ModifyManagedObjectStorageUserAccessKey(ctx, &request.ModifyManagedObjectStorageUserAccessKeyRequest{
		ServiceUUID: data.ServiceUUID.ValueString(),
		Username:    data.Username.ValueString(),
		AccessKeyID: data.AccessKeyID.ValueString(),
		Status:      upcloud.ManagedObjectStorageUserAccessKeyStatus(data.Status.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update managed object storage user access key",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setUserAccessKeyValues(ctx, &data, accessKey)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageUserAccessKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data userAccessKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var serviceUUID, username, accessKeyID string
	if err := utils.UnmarshalID(data.ID.ValueString(), &serviceUUID, &username, &accessKeyID); err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal managed object storage user access key ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if err := r.client.DeleteManagedObjectStorageUserAccessKey(ctx, &request.DeleteManagedObjectStorageUserAccessKeyRequest{
		ServiceUUID: serviceUUID,
		Username:    username,
		AccessKeyID: accessKeyID,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage user access key",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
}

func (r *managedObjectStorageUserAccessKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
