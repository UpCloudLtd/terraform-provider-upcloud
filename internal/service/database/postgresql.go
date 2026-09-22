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
	_ resource.Resource                = &postgresResource{}
	_ resource.ResourceWithConfigure   = &postgresResource{}
	_ resource.ResourceWithImportState = &postgresResource{}
	_ resource.ResourceWithModifyPlan  = &postgresResource{}
)

func NewPostgresResource() resource.Resource {
	return &postgresResource{}
}

type postgresResource struct {
	client *v9.ClientWithResponses
}

func (r *postgresResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_database_postgresql"
}

// Configure adds the provider configured client to the resource.
func (r *postgresResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	var diags diag.Diagnostics
	r.client, diags = utils.GetV9ClientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
}

type postgresModel struct {
	databaseCommonModel
	databasePlanModel

	SSLMode types.String `tfsdk:"sslmode"`
}

func (r *postgresResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: serviceDescription("PostgreSQL"),
		Attributes: map[string]schema.Attribute{
			"sslmode": schema.StringAttribute{
				MarkdownDescription: "SSL Connection Mode for PostgreSQL",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{},
	}

	defineCommonAttributesAndBlocks(&resp.Schema, upcloud.ManagedDatabaseServiceTypePostgreSQL)
}

func (r *postgresResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *postgresModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || plan == nil {
		return
	}

	var config postgresModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.PlanCompute.IsNull() || config.PlanCompute.IsUnknown() {
		return
	}

	var state *postgresModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || state == nil || !databaseComponentPlanChanged(&state.databasePlanModel, &plan.databasePlanModel) {
		return
	}

	plan.Plan = types.StringUnknown()
	plan.AdditionalDiskSpaceGiB = types.Int64Unknown()
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

func (r *postgresResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data postgresModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Type = types.StringValue(string(upcloud.ManagedDatabaseServiceTypePostgreSQL))

	db, diags := createDatabase(ctx, &data.databaseCommonModel, &data.databasePlanModel, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.SSLMode = types.StringValue(serviceURIParamString(db.ServiceUriParams, "ssl_mode"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *postgresResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data postgresModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	db, diags := readDatabase(ctx, &data.databaseCommonModel, &data.databasePlanModel, r.client, resp.State.RemoveResource)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || db == nil {
		return
	}

	data.SSLMode = types.StringValue(serviceURIParamString(db.ServiceUriParams, "ssl_mode"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *postgresResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var config, plan, state postgresModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	db, newVersion, d := updateDatabase(
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

	powered := db != nil && db.State != nil && *db.State == databaseStateRunning
	if newVersion != "" {
		resp.Diagnostics.Append(updateVersion(ctx, state.ID.ValueString(), newVersion, powered, r.client)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	db, diags := readDatabase(ctx, &plan.databaseCommonModel, &plan.databasePlanModel, r.client, resp.State.RemoveResource)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || db == nil {
		return
	}

	plan.SSLMode = types.StringValue(serviceURIParamString(db.ServiceUriParams, "ssl_mode"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *postgresResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data postgresModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	resp.Diagnostics.Append(deleteDatabase(ctx, r.client, data.ID.ValueString())...)
}

func (r *postgresResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
