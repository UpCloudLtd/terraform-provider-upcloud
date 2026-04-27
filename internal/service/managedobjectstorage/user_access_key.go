package managedobjectstorage

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
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

func NewUserAccessKeyResource() resource.Resource {
	return &managedObjectStorageUserAccessKeyResource{}
}

type managedObjectStorageUserAccessKeyResource struct {
	client *v9.ClientWithResponses
}

func (r *managedObjectStorageUserAccessKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage_user_access_key"
}

// Configure adds the provider configured client to the resource.
func (r *managedObjectStorageUserAccessKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
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
						string(v9.ObjectStorage2AccessKeyDetailResponseStatusActive),
						string(v9.ObjectStorage2AccessKeyDetailResponseStatusInactive),
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

func setUserAccessKeyValues(data *userAccessKeyModel, accessKey *v9.ObjectStorage2AccessKeyDetailResponse) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	if accessKey.AccessKeyId != nil {
		data.AccessKeyID = types.StringValue(*accessKey.AccessKeyId)
	}
	if accessKey.CreatedAt != nil {
		data.CreatedAt = types.StringValue(accessKey.CreatedAt.String())
	}
	if accessKey.LastUsedAt != nil {
		data.LastUsedAt = types.StringValue(accessKey.LastUsedAt.String())
	}
	if accessKey.Status != nil {
		data.Status = types.StringValue(string(*accessKey.Status))
	}

	return respDiagnostics
}

func (r *managedObjectStorageUserAccessKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data userAccessKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svcUUID, err := uuid.Parse(data.ServiceUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid service UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	apiResp, err := r.client.CreateObjectStorageAccessKeyWithResponse(ctx, svcUUID, data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage user access key",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.StatusCode() != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage user access key",
			objectStorageAPIErrorDetail(apiResp.ApplicationproblemJSONDefault, apiResp.Body),
		)
		return
	}
	if apiResp.JSON201 == nil {
		var dest v9.ObjectStorage2CreateAccessKey201
		if err := json.Unmarshal(apiResp.Body, &dest); err != nil {
			resp.Diagnostics.AddError(
				"Unable to read created managed object storage user access key",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
		apiResp.JSON201 = &dest
	}

	created := apiResp.JSON201
	accessKeyID := ""
	if created.AccessKeyId != nil {
		accessKeyID = *created.AccessKeyId
	}

	// Status can be set only after creation
	desiredStatus := v9.ObjectStorage2AccessKeyModifyStatus(data.Status.ValueString())
	if created.Status == nil || string(*created.Status) != data.Status.ValueString() {
		modResp, err := r.client.ModifyObjectStorageAccessKeyDetailsWithResponse(ctx, svcUUID, data.Username.ValueString(), accessKeyID, v9.ModifyObjectStorageAccessKeyDetailsJSONRequestBody{
			Status: &desiredStatus,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to set managed object storage user access key status",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
		if modResp.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError(
				"Unable to set managed object storage user access key status",
				objectStorageAPIErrorDetail(modResp.ApplicationproblemJSONDefault, modResp.Body),
			)
			return
		}
	}

	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.Username.ValueString(), accessKeyID))
	if created.SecretAccessKey != nil {
		data.SecretAccessKey = types.StringValue(*created.SecretAccessKey)
	}
	if created.AccessKeyId != nil {
		data.AccessKeyID = types.StringValue(*created.AccessKeyId)
	}
	if created.CreatedAt != nil {
		data.CreatedAt = types.StringValue(created.CreatedAt.String())
	}
	if created.LastUsedAt != nil {
		data.LastUsedAt = types.StringValue(created.LastUsedAt.String())
	}

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
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &username, &accessKeyID)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ServiceUUID = types.StringValue(serviceUUID)
	data.Username = types.StringValue(username)
	data.AccessKeyID = types.StringValue(accessKeyID)

	svcUUID, err := uuid.Parse(serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid service UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	apiResp, err := r.client.GetObjectStorageAccessKeyDetailsWithResponse(ctx, svcUUID, username, accessKeyID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object storage user access key details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to read managed object storage user access key details",
			objectStorageAPIErrorDetail(apiResp.ApplicationproblemJSONDefault, apiResp.Body),
		)
		return
	}
	if apiResp.JSON200 == nil {
		var dest v9.ObjectStorage2GetAccessKeyDetails200
		if err := json.Unmarshal(apiResp.Body, &dest); err != nil {
			resp.Diagnostics.AddError(
				"Unable to read managed object storage user access key details",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
		apiResp.JSON200 = &dest
	}

	resp.Diagnostics.Append(setUserAccessKeyValues(&data, apiResp.JSON200)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageUserAccessKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data userAccessKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svcUUID, err := uuid.Parse(data.ServiceUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid service UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	desiredStatus := v9.ObjectStorage2AccessKeyModifyStatus(data.Status.ValueString())
	apiResp, err := r.client.ModifyObjectStorageAccessKeyDetailsWithResponse(ctx, svcUUID, data.Username.ValueString(), data.AccessKeyID.ValueString(), v9.ModifyObjectStorageAccessKeyDetailsJSONRequestBody{
		Status: &desiredStatus,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update managed object storage user access key",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to update managed object storage user access key",
			objectStorageAPIErrorDetail(apiResp.ApplicationproblemJSONDefault, apiResp.Body),
		)
		return
	}
	if apiResp.JSON200 == nil {
		var dest v9.ObjectStorage2ModifyAccessKeyDetails200
		if err := json.Unmarshal(apiResp.Body, &dest); err != nil {
			resp.Diagnostics.AddError(
				"Unable to update managed object storage user access key",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
		apiResp.JSON200 = &dest
	}

	resp.Diagnostics.Append(setUserAccessKeyValues(&data, apiResp.JSON200)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageUserAccessKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data userAccessKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var serviceUUID, username, accessKeyID string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &username, &accessKeyID)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svcUUID, err := uuid.Parse(serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid service UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	apiResp, err := r.client.DeleteObjectStorageAccessKeyWithResponse(ctx, svcUUID, username, accessKeyID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage user access key",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.StatusCode() != http.StatusNoContent && apiResp.StatusCode() != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage user access key",
			objectStorageAPIErrorDetail(apiResp.ApplicationproblemJSONDefault, apiResp.Body),
		)
	}
}

func (r *managedObjectStorageUserAccessKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
