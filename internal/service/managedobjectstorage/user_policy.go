package managedobjectstorage

import (
	"context"

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
	_ resource.Resource                = &managedObjectStorageUserPolicyResource{}
	_ resource.ResourceWithConfigure   = &managedObjectStorageUserPolicyResource{}
	_ resource.ResourceWithImportState = &managedObjectStorageUserPolicyResource{}
)

func NewUserPolicyResource() resource.Resource {
	return &managedObjectStorageUserPolicyResource{}
}

type managedObjectStorageUserPolicyResource struct {
	client *service.Service
}

func (r *managedObjectStorageUserPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage_user_policy"
}

// Configure adds the provider configured client to the resource.
func (r *managedObjectStorageUserPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type userPolicyModel struct {
	Name        types.String `tfsdk:"name"`
	ID          types.String `tfsdk:"id"`
	ServiceUUID types.String `tfsdk:"service_uuid"`
	Username    types.String `tfsdk:"username"`
}

func (r *managedObjectStorageUserPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an UpCloud Managed Object Storage user policy attachment.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Policy name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "ID of the user. ID is in {object storage UUID}/{username}/{policy name} format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username.",
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

func (r *managedObjectStorageUserPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data userPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.Username.ValueString(), data.Name.ValueString()))

	apiReq := &request.AttachManagedObjectStorageUserPolicyRequest{
		ServiceUUID: data.ServiceUUID.ValueString(),
		Username:    data.Username.ValueString(),
		Name:        data.Name.ValueString(),
	}

	err := r.client.AttachManagedObjectStorageUserPolicy(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage user policy",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func policyExists(policies []upcloud.ManagedObjectStorageUserPolicy, name string) bool {
	for _, p := range policies {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (r *managedObjectStorageUserPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data userPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	var serviceUUID, username, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &username, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ServiceUUID = types.StringValue(serviceUUID)
	data.Username = types.StringValue(username)
	data.Name = types.StringValue(name)

	policies, err := r.client.GetManagedObjectStorageUserPolicies(ctx, &request.GetManagedObjectStorageUserPoliciesRequest{
		Username:    username,
		ServiceUUID: serviceUUID,
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read managed object storage user policies",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	if !policyExists(policies, name) {
		resp.State.RemoveResource(ctx)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageUserPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All configurable attributes require replace, so Update method is only required to satisfy the interface.
}

func (r *managedObjectStorageUserPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data userPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var serviceUUID, username, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &username, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DetachManagedObjectStorageUserPolicy(ctx, &request.DetachManagedObjectStorageUserPolicyRequest{
		ServiceUUID: serviceUUID,
		Username:    username,
		Name:        name,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage user policy",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *managedObjectStorageUserPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
