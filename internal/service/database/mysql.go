package database

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &mysqlResource{}
	_ resource.ResourceWithConfigure   = &mysqlResource{}
	_ resource.ResourceWithImportState = &mysqlResource{}
	_ resource.ResourceWithModifyPlan  = &mysqlResource{}
)

func NewMySQLResource() resource.Resource {
	return &mysqlResource{}
}

type mysqlResource struct {
	client *v9.ClientWithResponses
}

type mysqlModel struct {
	databaseCommonModel
	databasePlanModel
}

func (r *mysqlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_database_mysql"
}

// Configure adds the provider configured client to the resource.
func (r *mysqlResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	var diags diag.Diagnostics
	r.client, diags = utils.GetV9ClientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
}

func (r *mysqlResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: serviceDescription("MySQL"),
		Attributes:          map[string]schema.Attribute{},
		Blocks:              map[string]schema.Block{},
	}

	defineCommonAttributesAndBlocks(&resp.Schema, upcloud.ManagedDatabaseServiceTypeMySQL)
}

func (r *mysqlResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *mysqlModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || plan == nil {
		return
	}

	var config mysqlModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.PlanCompute.IsNull() || config.PlanCompute.IsUnknown() {
		return
	}

	var state *mysqlModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || state == nil || !databaseComponentPlanChanged(&state.databasePlanModel, &plan.databasePlanModel) {
		return
	}

	plan.Plan = types.StringUnknown()
	plan.AdditionalDiskSpaceGiB = types.Int64Unknown()
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

func (r *mysqlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data mysqlModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Type = types.StringValue(string(upcloud.ManagedDatabaseServiceTypeMySQL))

	_, diags := createDatabase(ctx, &data.databaseCommonModel, &data.databasePlanModel, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *mysqlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data mysqlModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	db, diags := readDatabase(ctx, &data.databaseCommonModel, &data.databasePlanModel, r.client, resp.State.RemoveResource)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || db == nil {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *mysqlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var config, plan, state mysqlModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, _, d := updateDatabase(
		ctx,
		&state.databaseCommonModel,
		&plan.databaseCommonModel,
		&config.databaseCommonModel,
		&state.databasePlanModel,
		&plan.databasePlanModel,
		&config.databasePlanModel,
		r.client,
	)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resolveUnknownDatabasePlanComponents(&plan.databasePlanModel, &state.databasePlanModel)
	db, diags := readDatabase(ctx, &plan.databaseCommonModel, &plan.databasePlanModel, r.client, resp.State.RemoveResource)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || db == nil {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mysqlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data mysqlModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	resp.Diagnostics.Append(deleteDatabase(ctx, r.client, data.ID.ValueString())...)
}

func (r *mysqlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
