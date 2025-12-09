package database

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &mysqlResource{}
	_ resource.ResourceWithConfigure   = &mysqlResource{}
	_ resource.ResourceWithImportState = &mysqlResource{}
)

func NewMySQLResource() resource.Resource {
	return &mysqlResource{}
}

type mysqlResource struct {
	client *service.Service
}

func (r *mysqlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_database_mysql"
}

// Configure adds the provider configured client to the resource.
func (r *mysqlResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

func (r *mysqlResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: serviceDescription("MySQL"),
		Attributes:          map[string]schema.Attribute{},
		Blocks:              map[string]schema.Block{},
	}

	defineCommonAttributesAndBlocks(&resp.Schema, upcloud.ManagedDatabaseServiceTypeMySQL)
}

func (r *mysqlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data databaseCommonModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Type = types.StringValue(string(upcloud.ManagedDatabaseServiceTypeMySQL))

	_, diags := createDatabase(ctx, &data, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *mysqlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data databaseCommonModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, diags := readDatabase(ctx, &data, r.client, resp.State.RemoveResource)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *mysqlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state databaseCommonModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, _, d := updateDatabase(ctx, &state, &plan, r.client)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, diags := readDatabase(ctx, &plan, r.client, resp.State.RemoveResource)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mysqlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data databaseCommonModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if err := r.client.DeleteManagedDatabase(ctx, &request.DeleteManagedDatabaseRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed database",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	resp.Diagnostics.Append(waitForDatabaseToBeDeletedDiags(ctx, r.client, data.ID.ValueString())...)
}

func (r *mysqlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
