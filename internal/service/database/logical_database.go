package database

import (
	"context"
	"fmt"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const localeError = "locale must be in form en_US.UTF8 (language_TERRITORY.CODEPOINT)"

var localeRegExp = regexp.MustCompile(`^[a-z]{2}_[A-Z]{2}\.[A-z0-9-]+$`)

var (
	_ resource.Resource                = &logicalDatabaseResource{}
	_ resource.ResourceWithConfigure   = &logicalDatabaseResource{}
	_ resource.ResourceWithImportState = &logicalDatabaseResource{}
)

func NewLogicalDatabaseResource() resource.Resource {
	return &logicalDatabaseResource{}
}

type logicalDatabaseResource struct {
	client *service.Service
}

func (r *logicalDatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_database_logical_database"
}

// Configure adds the provider configured client to the resource.
func (r *logicalDatabaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type logicalDatabaseModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Service      types.String `tfsdk:"service"`
	CharacterSet types.String `tfsdk:"character_set"`
	Collation    types.String `tfsdk:"collation"`
}

func (r *logicalDatabaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `This resource represents a logical database in managed database.`,
		Attributes: map[string]schema.Attribute{
			"service": schema.StringAttribute{
				Description: "UUID of the service to which this logical database belongs.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "ID of the logical database. ID is in {service UUID}/{database name} format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the logical database",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"character_set": schema.StringAttribute{
				Description: "Default character set for the database (LC_CTYPE), PostgreSQL only.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(localeRegExp, localeError),
				},
			},
			"collation": schema.StringAttribute{
				Description: "Default collation for the database (LC_COLLATE), PostgreSQL only.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(localeRegExp, localeError),
				},
			},
		},
	}
}

func (r *logicalDatabaseResource) setLogicalDatabaseValues(ctx context.Context, data *logicalDatabaseModel, svcUUID, ldbName string) error {
	ldbs, err := r.client.GetManagedDatabaseLogicalDatabases(ctx, &request.GetManagedDatabaseLogicalDatabasesRequest{
		ServiceUUID: svcUUID,
	})
	if err != nil {
		return err
	}

	for _, ldb := range ldbs {
		if ldb.Name == ldbName {
			data.Name = types.StringValue(ldb.Name)
			data.CharacterSet = types.StringValue(ldb.LCCType)
			data.Collation = types.StringValue(ldb.LCCollate)
			return nil
		}
	}
	return fmt.Errorf("logical database %s not found", ldbName)
}

func (r *logicalDatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data logicalDatabaseModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	uuid := data.Service.ValueString()

	serviceDetails, err := r.client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: uuid})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get managed database service details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if !serviceDetails.Powered {
		resp.Diagnostics.AddError(
			"Unable to create logical database",
			fmt.Sprintf("cannot create a logical database while managed database %v (%v) is powered off", serviceDetails.Name, uuid),
		)
		return
	}

	if (data.CharacterSet.ValueString() != "" || data.Collation.ValueString() != "") && serviceDetails.Type != upcloud.ManagedDatabaseServiceTypePostgreSQL {
		resp.Diagnostics.AddError(
			"Invalid character set or collation",
			"Setting character_set or collation is only possible for PostgreSQL service",
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(uuid, data.Name.ValueString()))

	name := data.Name.ValueString()
	apiReq := &request.CreateManagedDatabaseLogicalDatabaseRequest{
		ServiceUUID: uuid,
		Name:        name,
		LCCollate:   data.Collation.ValueString(),
		LCCType:     data.CharacterSet.ValueString(),
	}

	_, err = r.client.CreateManagedDatabaseLogicalDatabase(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed database logical database",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	err = r.setLogicalDatabaseValues(ctx, &data, uuid, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read logical database details after creation",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *logicalDatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data logicalDatabaseModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	var uuid, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &uuid, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Service = types.StringValue(uuid)

	err := r.setLogicalDatabaseValues(ctx, &data, uuid, name)
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read logical databases details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *logicalDatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No operation, all changes require replace.
}

func (r *logicalDatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data logicalDatabaseModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var uuid, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &uuid, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteManagedDatabaseLogicalDatabase(ctx, &request.DeleteManagedDatabaseLogicalDatabaseRequest{
		ServiceUUID: uuid,
		Name:        name,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete logical database",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *logicalDatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
