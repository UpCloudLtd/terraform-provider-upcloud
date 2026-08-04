package database

import (
	"context"
	"fmt"
	"net/http"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &connectionPoolResource{}
	_ resource.ResourceWithConfigure   = &connectionPoolResource{}
	_ resource.ResourceWithImportState = &connectionPoolResource{}
)

func NewConnectionPoolResource() resource.Resource {
	return &connectionPoolResource{}
}

type connectionPoolResource struct {
	client *v9.ClientWithResponses
}

func (r *connectionPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_database_connection_pool"
}

// Configure adds the provider configured client to the resource.
func (r *connectionPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

type connectionPoolModel struct {
	ID            types.String `tfsdk:"id"`
	Service       types.String `tfsdk:"service"`
	Database      types.String `tfsdk:"database"`
	Name          types.String `tfsdk:"name"`
	Mode          types.String `tfsdk:"mode"`
	Size          types.Int32  `tfsdk:"size"`
	Username      types.String `tfsdk:"username"`
	ConnectionURI types.String `tfsdk:"connection_uri"`
}

func (r *connectionPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `This resource represents a connection pool in a managed database (PostgreSQL).`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the connection pool. ID is in {service UUID}/{connection pool name} format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service": schema.StringAttribute{
				Description: "UUID of the service to which this connection pool belongs.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database": schema.StringAttribute{
				Description: "Name of the database.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Connection pool name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mode": schema.StringAttribute{
				Description: "Connection pool mode.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("transaction", "session", "statement"),
				},
			},
			"size": schema.Int32Attribute{
				Description: "Connection pool size.",
				Required:    true,
				Validators: []validator.Int32{
					int32validator.Between(1, 10000),
				},
			},
			"username": schema.StringAttribute{
				Description: "Service username. If not set, all users are allowed to access this connection pool.",
				Optional:    true,
			},
			"connection_uri": schema.StringAttribute{
				Description: "Connection URI for the connection pool.",
				Computed:    true,
			},
		},
	}
}

func setConnectionPoolValues(data *connectionPoolModel, pool *v9.DatabaseConnectionPoolResponse) {
	data.Database = types.StringPointerValue(pool.Database)
	data.Name = types.StringPointerValue(pool.PoolName)
	data.Mode = types.StringPointerValue(pool.PoolMode)
	data.Size = types.Int32PointerValue(pool.PoolSize)
	data.Username = types.StringPointerValue(pool.Username)
	data.ConnectionURI = types.StringPointerValue(pool.ConnectionUri)
}

func (r *connectionPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionPoolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svcUUID, err := uuid.Parse(data.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse database service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(svcUUID.String(), data.Name.ValueString()))

	apiReq := v9.CreateDatabaseConnectionPoolJSONRequestBody{
		Database: data.Database.ValueString(),
		PoolName: data.Name.ValueString(),
		PoolMode: v9.DatabaseConnectionPoolCreatePoolMode(data.Mode.ValueString()),
		PoolSize: data.Size.ValueInt32(),
		Username: utils.ValueStringOrNil(data.Username),
	}

	apiResp, err := r.client.CreateDatabaseConnectionPoolWithResponse(ctx, svcUUID, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create connection pool",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if apiResp.StatusCode() != http.StatusCreated || apiResp.JSON201 == nil {
		resp.Diagnostics.AddError(
			"Unable to create connection pool",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.Service.ValueString(), data.Name.ValueString()))

	setConnectionPoolValues(&data, apiResp.JSON201)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	var uuidStr, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &uuidStr, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Service = types.StringValue(uuidStr)

	svcUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse database service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	apiResp, err := r.client.GetDatabaseConnectionPoolWithResponse(ctx, svcUUID, name)
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read managed database connection pool details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	if apiResp.StatusCode() != http.StatusOK || apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to read connection pool",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}

	setConnectionPoolValues(&data, apiResp.JSON200)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data connectionPoolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var uuidStr, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &uuidStr, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Service = types.StringValue(uuidStr)

	svcUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse database service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	apiReq := v9.ModifyDatabaseConnectionPoolJSONRequestBody{
		Database: utils.ValueStringOrNil(data.Database),
		PoolMode: (*v9.DatabaseConnectionPoolModifyPoolMode)(utils.ValueStringOrNil(data.Mode)),
		PoolSize: utils.ValueInt32OrNil(data.Size),
		Username: utils.ValueStringOrNil(data.Username),
	}

	apiResp, err := r.client.ModifyDatabaseConnectionPoolWithResponse(ctx, svcUUID, name, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify connection pool",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if apiResp.StatusCode() != http.StatusOK || apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to modify connection pool",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}

	setConnectionPoolValues(&data, apiResp.JSON200)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var serviceUUID, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svcUUID, err := uuid.Parse(serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse database service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	apiResp, err := r.client.DeleteDatabaseConnectionPoolWithResponse(ctx, svcUUID, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete connection pool",
			utils.ErrorDiagnosticDetail(err),
		)
	} else if apiResp.StatusCode() < 200 || apiResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError(
			"Unable to delete connection pool",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
	}
}

func (r *connectionPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
