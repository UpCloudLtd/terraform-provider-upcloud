package database

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &databaseUserResource{}
	_ resource.ResourceWithConfigure   = &databaseUserResource{}
	_ resource.ResourceWithImportState = &databaseUserResource{}
)

func NewUserResource() resource.Resource {
	return &databaseUserResource{}
}

type databaseUserResource struct {
	client *service.Service
}

func (r *databaseUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_database_user"
}

// Configure adds the provider configured client to the resource.
func (r *databaseUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type databaseUserModel struct {
	ID                      types.String `tfsdk:"id"`
	Service                 types.String `tfsdk:"service"`
	Username                types.String `tfsdk:"username"`
	Password                types.String `tfsdk:"password"`
	Type                    types.String `tfsdk:"type"`
	Authentication          types.String `tfsdk:"authentication"`
	PgAccessControl         types.List   `tfsdk:"pg_access_control"`
	ValkeyAccessControl     types.List   `tfsdk:"valkey_access_control"`
	OpensearchAccessControl types.List   `tfsdk:"opensearch_access_control"`
}

type pgAccessControlModel struct {
	AllowReplication types.Bool `tfsdk:"allow_replication"`
}

type valkeyAccessControlModel struct {
	Categories types.List `tfsdk:"categories"`
	Channels   types.List `tfsdk:"channels"`
	Commands   types.List `tfsdk:"commands"`
	Keys       types.List `tfsdk:"keys"`
}

type opensearchAccessControlModel struct {
	Rules types.List `tfsdk:"rules"`
}

type opensearchAccessControlRuleModel struct {
	Index      types.String `tfsdk:"index"`
	Permission types.String `tfsdk:"permission"`
}

func (r *databaseUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	emptyStringList := types.ListValueMust(types.StringType, []attr.Value{})

	resp.Schema = schema.Schema{
		MarkdownDescription: `This resource represents a user in managed database.`,
		Attributes: map[string]schema.Attribute{
			"service": schema.StringAttribute{
				Description: "Service's UUID for which this user belongs to",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "ID of the user. ID is in {service UUID}/{username} format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Name of the database user",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Description: "Password for the database user. Defaults to a random value",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(8, 256),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of the user. Only normal type users can be created",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"authentication": schema.StringAttribute{
				Description: "MySQL only, authentication type.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(upcloud.ManagedDatabaseUserAuthenticationCachingSHA2Password),
						string(upcloud.ManagedDatabaseUserAuthenticationMySQLNativePassword),
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"pg_access_control": schema.ListNestedBlock{
				Description: "PostgreSQL access control object.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"allow_replication": schema.BoolAttribute{
							Description: "Grant replication privilege",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 1),
					listvalidator.ConflictsWith(
						path.MatchRoot("opensearch_access_control"),
						path.MatchRoot("valkey_access_control"),
					),
				},
			},
			"valkey_access_control": schema.ListNestedBlock{
				Description: "Valkey access control object.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"categories": schema.ListAttribute{
							Description: "Set access control to all commands in specified categories.",
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
							Default:     listdefault.StaticValue(emptyStringList),
						},
						"channels": schema.ListAttribute{
							Description: "Set access control to Pub/Sub channels.",
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
							Default:     listdefault.StaticValue(emptyStringList),
						},
						"commands": schema.ListAttribute{
							Description: "Set access control to commands.",
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
							Default:     listdefault.StaticValue(emptyStringList),
						},
						"keys": schema.ListAttribute{
							Description: "Set access control to keys.",
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
							Default:     listdefault.StaticValue(emptyStringList),
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 1),
					listvalidator.ConflictsWith(
						path.MatchRoot("opensearch_access_control"),
					),
				},
			},
			"opensearch_access_control": schema.ListNestedBlock{
				Description: "OpenSearch access control object.",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"rules": schema.ListNestedBlock{
							Description: "Set user access control rules.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"index": schema.StringAttribute{
										Description: "Set index name, pattern or top level API.",
										Required:    true,
									},
									"permission": schema.StringAttribute{
										Description: "Set permission access.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(
												string(upcloud.ManagedDatabaseUserOpenSearchAccessControlRulePermissionAdmin),
												string(upcloud.ManagedDatabaseUserOpenSearchAccessControlRulePermissionDeny),
												string(upcloud.ManagedDatabaseUserOpenSearchAccessControlRulePermissionRead),
												string(upcloud.ManagedDatabaseUserOpenSearchAccessControlRulePermissionReadWrite),
												string(upcloud.ManagedDatabaseUserOpenSearchAccessControlRulePermissionWrite),
											),
										},
									},
								},
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 1),
				},
			},
		},
	}
}

func listValuePointerMust(s *[]string) types.List {
	if s == nil {
		return types.ListNull(types.StringType)
	}

	vals := make([]attr.Value, len(*s))
	for i, v := range *s {
		vals[i] = types.StringValue(v)
	}

	return types.ListValueMust(types.StringType, vals)
}

func setDatabaseUserValues(ctx context.Context, data *databaseUserModel, user *upcloud.ManagedDatabaseUser) diag.Diagnostics {
	var respDiagnostics, d diag.Diagnostics

	isImport := data.Username.ValueString() == ""

	data.Username = types.StringValue(user.Username)
	data.Password = types.StringValue(user.Password)
	data.Type = types.StringValue(string(user.Type))

	if !data.Authentication.IsNull() || (isImport && user.Authentication != "") {
		data.Authentication = types.StringValue(string(user.Authentication))
	}

	if (!data.PgAccessControl.IsNull() || isImport) && user.PGAccessControl != nil {
		pgData := pgAccessControlModel{
			AllowReplication: types.BoolPointerValue(user.PGAccessControl.AllowReplication),
		}
		data.PgAccessControl, d = types.ListValueFrom(ctx, data.PgAccessControl.ElementType(ctx), []pgAccessControlModel{pgData})
		respDiagnostics.Append(d...)
	}

	if (!data.ValkeyAccessControl.IsNull() || isImport) && user.ValkeyAccessControl != nil {
		valkeyData := valkeyAccessControlModel{
			Categories: listValuePointerMust(user.ValkeyAccessControl.Categories),
			Channels:   listValuePointerMust(user.ValkeyAccessControl.Channels),
			Commands:   listValuePointerMust(user.ValkeyAccessControl.Commands),
			Keys:       listValuePointerMust(user.ValkeyAccessControl.Keys),
		}
		data.ValkeyAccessControl, d = types.ListValueFrom(ctx, data.ValkeyAccessControl.ElementType(ctx), []valkeyAccessControlModel{valkeyData})
		respDiagnostics.Append(d...)
	}

	if (!data.OpensearchAccessControl.IsNull() || isImport) && user.OpenSearchAccessControl != nil && user.OpenSearchAccessControl.Rules != nil {
		rulesData := make([]opensearchAccessControlRuleModel, len(*user.OpenSearchAccessControl.Rules))
		for i, r := range *user.OpenSearchAccessControl.Rules {
			rulesData[i] = opensearchAccessControlRuleModel{
				Index:      types.StringValue(r.Index),
				Permission: types.StringValue(string(r.Permission)),
			}
		}

		rules, d := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"index":      types.StringType,
				"permission": types.StringType,
			},
		}, rulesData)
		respDiagnostics.Append(d...)

		data.OpensearchAccessControl, d = types.ListValueFrom(ctx, data.OpensearchAccessControl.ElementType(ctx), []opensearchAccessControlModel{{Rules: rules}})
		respDiagnostics.Append(d...)
	}

	return respDiagnostics
}

func getPgAccessControl(ctx context.Context, data *databaseUserModel) (*upcloud.ManagedDatabaseUserPGAccessControl, diag.Diagnostics) {
	if data.PgAccessControl.IsUnknown() {
		return nil, nil
	}

	if data.PgAccessControl.IsNull() {
		return &upcloud.ManagedDatabaseUserPGAccessControl{
			AllowReplication: upcloud.BoolPtr(true),
		}, nil
	}

	var pgAccessControl []pgAccessControlModel
	d := data.PgAccessControl.ElementsAs(ctx, &pgAccessControl, false)
	if d.HasError() {
		return nil, d
	}

	if len(pgAccessControl) == 0 {
		return nil, nil
	}

	return &upcloud.ManagedDatabaseUserPGAccessControl{
		AllowReplication: upcloud.BoolPtr(*pgAccessControl[0].AllowReplication.ValueBoolPointer()),
	}, nil
}

func asStringSlicePointer(ctx context.Context, list types.List) (*[]string, diag.Diagnostics) {
	if list.IsUnknown() {
		return nil, nil
	}

	stringSlice := []string{}

	if list.IsNull() {
		return &stringSlice, nil
	}

	d := list.ElementsAs(ctx, &stringSlice, false)
	return &stringSlice, d
}

func getValkeyAccessControl(ctx context.Context, data *databaseUserModel) (*upcloud.ManagedDatabaseUserValkeyAccessControl, diag.Diagnostics) {
	if data.ValkeyAccessControl.IsUnknown() {
		return nil, nil
	}

	if data.ValkeyAccessControl.IsNull() {
		return &upcloud.ManagedDatabaseUserValkeyAccessControl{
			Categories: &[]string{},
			Channels:   &[]string{},
			Commands:   &[]string{},
			Keys:       &[]string{},
		}, nil
	}

	var valkeyAccessControl []valkeyAccessControlModel
	d := data.ValkeyAccessControl.ElementsAs(ctx, &valkeyAccessControl, false)
	if d.HasError() {
		return nil, d
	}

	if len(valkeyAccessControl) == 0 {
		return nil, nil
	}

	var diags diag.Diagnostics

	categories, d := asStringSlicePointer(ctx, valkeyAccessControl[0].Categories)
	diags.Append(d...)

	channels, d := asStringSlicePointer(ctx, valkeyAccessControl[0].Channels)
	diags.Append(d...)

	commands, d := asStringSlicePointer(ctx, valkeyAccessControl[0].Commands)
	diags.Append(d...)

	keys, d := asStringSlicePointer(ctx, valkeyAccessControl[0].Keys)
	diags.Append(d...)

	return &upcloud.ManagedDatabaseUserValkeyAccessControl{
		Categories: categories,
		Channels:   channels,
		Commands:   commands,
		Keys:       keys,
	}, diags
}

func getOpensearchAccessControl(ctx context.Context, data *databaseUserModel) (*upcloud.ManagedDatabaseUserOpenSearchAccessControl, diag.Diagnostics) {
	if data.OpensearchAccessControl.IsUnknown() {
		return nil, nil
	}

	if data.OpensearchAccessControl.IsNull() {
		return &upcloud.ManagedDatabaseUserOpenSearchAccessControl{
			Rules: &[]upcloud.ManagedDatabaseUserOpenSearchAccessControlRule{},
		}, nil
	}

	var opensearchAccessControl []opensearchAccessControlModel
	d := data.OpensearchAccessControl.ElementsAs(ctx, &opensearchAccessControl, false)
	if d.HasError() {
		return nil, d
	}

	if len(opensearchAccessControl) == 0 {
		return nil, nil
	}

	var rulesData []opensearchAccessControlRuleModel
	d = opensearchAccessControl[0].Rules.ElementsAs(ctx, &rulesData, false)
	if d.HasError() {
		return nil, d
	}

	rules := make([]upcloud.ManagedDatabaseUserOpenSearchAccessControlRule, len(rulesData))

	for i, r := range rulesData {
		rules[i] = upcloud.ManagedDatabaseUserOpenSearchAccessControlRule{
			Index:      r.Index.ValueString(),
			Permission: upcloud.ManagedDatabaseUserOpenSearchAccessControlRulePermission(r.Permission.ValueString()),
		}
	}

	return &upcloud.ManagedDatabaseUserOpenSearchAccessControl{
		Rules: &rules,
	}, nil
}

func checkDatabaseIsRunning(ctx context.Context, client *service.Service, uuid, action string) (diags diag.Diagnostics) {
	db, err := client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: uuid})
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Unable to read managed database details during user %s", action),
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if !db.Powered {
		diags.AddError(
			"Managed database service is not powered on",
			fmt.Sprintf("Managed database service with UUID %s must be powered on to %s users", uuid, action),
		)
		return
	}

	_, err = client.WaitForManagedDatabaseState(ctx, &request.WaitForManagedDatabaseStateRequest{
		UUID:         uuid,
		DesiredState: upcloud.ManagedDatabaseStateRunning,
	})
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Error waiting for managed database to be running during user %s", action),
			utils.ErrorDiagnosticDetail(err),
		)
	}
	return
}

func (r *databaseUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data databaseUserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	uuid := data.Service.ValueString()

	resp.Diagnostics.Append(checkDatabaseIsRunning(ctx, r.client, uuid, "create")...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(utils.MarshalID(uuid, data.Username.ValueString()))

	apiReq := &request.CreateManagedDatabaseUserRequest{
		ServiceUUID:    uuid,
		Username:       data.Username.ValueString(),
		Password:       data.Password.ValueString(),
		Authentication: upcloud.ManagedDatabaseUserAuthenticationType(data.Authentication.ValueString()),
	}

	pgAccessControl, d := getPgAccessControl(ctx, &data)
	apiReq.PGAccessControl = pgAccessControl
	resp.Diagnostics.Append(d...)

	valkeyAccessControl, d := getValkeyAccessControl(ctx, &data)
	apiReq.ValkeyAccessControl = valkeyAccessControl
	resp.Diagnostics.Append(d...)

	opensearchAccessControl, d := getOpensearchAccessControl(ctx, &data)
	apiReq.OpenSearchAccessControl = opensearchAccessControl
	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.CreateManagedDatabaseUser(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed database user",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	setDatabaseUserValues(ctx, &data, user)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *databaseUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data databaseUserModel
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

	user, err := r.client.GetManagedDatabaseUser(ctx, &request.GetManagedDatabaseUserRequest{
		ServiceUUID: uuid,
		Username:    name,
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read managed database user details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	setDatabaseUserValues(ctx, &data, user)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *databaseUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state databaseUserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var uuid, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &uuid, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(checkDatabaseIsRunning(ctx, r.client, uuid, "update")...)
	if resp.Diagnostics.HasError() {
		return
	}

	userReq := &request.ModifyManagedDatabaseUserRequest{
		ServiceUUID:    uuid,
		Username:       name,
		Password:       data.Password.ValueString(),
		Authentication: upcloud.ManagedDatabaseUserAuthenticationType(data.Authentication.ValueString()),
	}

	user, err := r.client.ModifyManagedDatabaseUser(ctx, userReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify managed database user",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	acHasChanges := false
	acReq := &request.ModifyManagedDatabaseUserAccessControlRequest{
		ServiceUUID: uuid,
		Username:    name,
	}

	if !data.PgAccessControl.Equal(state.PgAccessControl) {
		acHasChanges = true
		pgAccessControl, d := getPgAccessControl(ctx, &data)
		acReq.PGAccessControl = pgAccessControl
		resp.Diagnostics.Append(d...)
	}

	if !data.ValkeyAccessControl.Equal(state.ValkeyAccessControl) {
		acHasChanges = true
		valkeyAccessControl, d := getValkeyAccessControl(ctx, &data)
		acReq.ValkeyAccessControl = valkeyAccessControl
		resp.Diagnostics.Append(d...)
	}

	if !data.OpensearchAccessControl.Equal(state.OpensearchAccessControl) {
		acHasChanges = true
		opensearchAccessControl, d := getOpensearchAccessControl(ctx, &data)
		acReq.OpenSearchAccessControl = opensearchAccessControl
		resp.Diagnostics.Append(d...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if acHasChanges {
		ac, err := r.client.ModifyManagedDatabaseUserAccessControl(ctx, acReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update managed database user access control",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}

		user.PGAccessControl = ac.PGAccessControl
		user.ValkeyAccessControl = ac.ValkeyAccessControl
		user.OpenSearchAccessControl = ac.OpenSearchAccessControl
	}

	resp.Diagnostics.Append(setDatabaseUserValues(ctx, &data, user)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *databaseUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data databaseUserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var uuid, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &uuid, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(checkDatabaseIsRunning(ctx, r.client, uuid, "delete")...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteManagedDatabaseUser(ctx, &request.DeleteManagedDatabaseUserRequest{
		ServiceUUID: uuid,
		Username:    name,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete database user",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *databaseUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
