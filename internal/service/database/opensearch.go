package database

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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &opensearchResource{}
	_ resource.ResourceWithConfigure   = &opensearchResource{}
	_ resource.ResourceWithImportState = &opensearchResource{}
)

func NewOpenSearchResource() resource.Resource {
	return &opensearchResource{}
}

type opensearchResource struct {
	client *service.Service
}

func (r *opensearchResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_database_opensearch"
}

// Configure adds the provider configured client to the resource.
func (r *opensearchResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type opensearchModel struct {
	databaseCommonModel

	AccessControl         types.Bool `tfsdk:"access_control"`
	ExtendedAccessControl types.Bool `tfsdk:"extended_access_control"`
}

func (r *opensearchResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: serviceDescription("OpenSearch"),
		Attributes: map[string]schema.Attribute{
			"access_control": schema.BoolAttribute{
				MarkdownDescription: "Enables users access control for OpenSearch service. User access control rules will only be enforced if this attribute is enabled.",
				Computed:            true,
				Optional:            true,
			},
			"extended_access_control": schema.BoolAttribute{
				MarkdownDescription: "Grant access to top-level `_mget`, `_msearch` and `_bulk` APIs. Users are limited to perform operations on indices based on the user-specific access control rules.",
				Computed:            true,
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{},
	}

	defineCommonAttributesAndBlocks(&resp.Schema, upcloud.ManagedDatabaseServiceTypeOpenSearch)
}

func updateAccessControlIfNeeded(ctx context.Context, client *service.Service, state, plan opensearchModel) (diags diag.Diagnostics) {
	if !state.AccessControl.Equal(plan.AccessControl) || !state.ExtendedAccessControl.Equal(plan.ExtendedAccessControl) {
		aclReq := request.ModifyManagedDatabaseAccessControlRequest{
			ServiceUUID:         plan.ID.ValueString(),
			ACLsEnabled:         upcloud.BoolPtr(plan.AccessControl.ValueBool()),
			ExtendedACLsEnabled: upcloud.BoolPtr(plan.ExtendedAccessControl.ValueBool()),
		}
		_, err := client.ModifyManagedDatabaseAccessControl(ctx, &aclReq)
		if err != nil {
			diags.AddError(
				"Unable to set OpenSearch access control settings",
				utils.ErrorDiagnosticDetail(err),
			)
			return diags
		}
	}

	return diags
}

func (r *opensearchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data opensearchModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Type = types.StringValue(string(upcloud.ManagedDatabaseServiceTypeOpenSearch))

	_, diags := createDatabase(ctx, &data.databaseCommonModel, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(updateAccessControlIfNeeded(ctx, r.client, opensearchModel{}, data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *opensearchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data opensearchModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, diags := readDatabase(ctx, &data.databaseCommonModel, r.client, resp.State.RemoveResource)
	resp.Diagnostics.Append(diags...)

	acl, err := r.client.GetManagedDatabaseAccessControl(ctx, &request.GetManagedDatabaseAccessControlRequest{
		ServiceUUID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read OpenSearch access control settings",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.AccessControl = types.BoolPointerValue(acl.ACLsEnabled)
	data.ExtendedAccessControl = types.BoolPointerValue(acl.ExtendedACLsEnabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *opensearchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state opensearchModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(updateAccessControlIfNeeded(ctx, r.client, state, plan)...)

	_, _, d := updateDatabase(ctx, &state.databaseCommonModel, &plan.databaseCommonModel, r.client)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, diags := readDatabase(ctx, &plan.databaseCommonModel, r.client, resp.State.RemoveResource)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *opensearchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data opensearchModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if err := r.client.DeleteManagedDatabase(ctx, &request.DeleteManagedDatabaseRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed database",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	resp.Diagnostics.Append(waitForDatabaseToBeDeleted(ctx, r.client, data.ID.ValueString())...)
}

func (r *opensearchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
