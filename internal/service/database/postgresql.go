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
	_ resource.Resource                = &postgresResource{}
	_ resource.ResourceWithConfigure   = &postgresResource{}
	_ resource.ResourceWithImportState = &postgresResource{}
)

func NewPostgresResource() resource.Resource {
	return &postgresResource{}
}

type postgresResource struct {
	client *service.Service
}

func (r *postgresResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_database_postgresql"
}

// Configure adds the provider configured client to the resource.
func (r *postgresResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type postgresModel struct {
	databaseCommonModel

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

func (r *postgresResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data postgresModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Type = types.StringValue(string(upcloud.ManagedDatabaseServiceTypePostgreSQL))

	db, diags := createDatabase(ctx, &data.databaseCommonModel, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.SSLMode = types.StringValue(db.ServiceURIParams.SSLMode)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *postgresResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data postgresModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	db, diags := readDatabase(ctx, &data.databaseCommonModel, r.client, resp.State.RemoveResource)
	resp.Diagnostics.Append(diags...)

	data.SSLMode = types.StringValue(db.ServiceURIParams.SSLMode)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *postgresResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Not implemented",
		"Updating PostgreSQL managed databases is not yet implemented.",
	)
}

func (r *postgresResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data postgresModel
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

func (r *postgresResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
