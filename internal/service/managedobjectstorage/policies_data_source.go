package managedobjectstorage

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPoliciesDataSource() datasource.DataSource {
	return &managedObjectStoragePoliciesDataSource{}
}

var (
	_ datasource.DataSource              = &managedObjectStoragePoliciesDataSource{}
	_ datasource.DataSourceWithConfigure = &managedObjectStoragePoliciesDataSource{}
)

type managedObjectStoragePoliciesDataSource struct {
	client *v9.ClientWithResponses
}

func (d *managedObjectStoragePoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage_policies"
}

func (d *managedObjectStoragePoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

type managedObjectStoragePoliciesModel struct {
	Policies    []managedObjectStoragePolicyModel `tfsdk:"policies"`
	ID          types.String                      `tfsdk:"id"`
	ServiceUUID types.String                      `tfsdk:"service_uuid"`
}

type managedObjectStoragePolicyModel struct {
	ARN              types.String `tfsdk:"arn"`
	AttachmentCount  types.Int64  `tfsdk:"attachment_count"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DefaultVersionID types.String `tfsdk:"default_version_id"`
	Description      types.String `tfsdk:"description"`
	Document         types.String `tfsdk:"document"`
	Name             types.String `tfsdk:"name"`
	ServiceUUID      types.String `tfsdk:"service_uuid"`
	System           types.Bool   `tfsdk:"system"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (d *managedObjectStoragePoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Policies available for a Managed Object Storage resource. See `managed_object_storage_user_policy` for attaching to a user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:           true,
				Description:        "The ID of this resource (same as `service_uuid`)",
				DeprecationMessage: "Contains the same value as `service_uuid`. Use `service_uuid` instead.",
			},
			"policies": schema.SetNestedAttribute{
				Description: "Policies.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"arn": schema.StringAttribute{
							Description: "Policy ARN.",
							Computed:    true,
						},
						"attachment_count": schema.Int64Attribute{
							Description: "Number of attachments.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Creation time.",
							Computed:    true,
						},
						"default_version_id": schema.StringAttribute{
							Description: "Default version ID.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Policy description.",
							Computed:    true,
						},
						"document": schema.StringAttribute{
							Description: "Policy document.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Policy name.",
							Computed:    true,
						},
						"service_uuid": schema.StringAttribute{
							Description: "Service UUID.",
							Computed:    true,
						},
						"system": schema.BoolAttribute{
							Description: "Whether the policy is a system policy.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "Last updated time.",
							Computed:    true,
						},
					},
				},
			},
			"service_uuid": schema.StringAttribute{
				Required:    true,
				Description: "Service UUID.",
			},
		},
	}
}

func (d *managedObjectStoragePoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data managedObjectStoragePoliciesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	svcUUID, err := uuid.Parse(data.ServiceUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object-storage policies",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	apiResp, err := d.client.ListObjectStoragePoliciesWithResponse(ctx, svcUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object-storage policies",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object-storage policies",
			utils.ErrorDiagnosticDetail(fmt.Errorf("unexpected response: %s", apiResp.HTTPResponse.Status)),
		)
		return
	}

	policies := *apiResp.JSON200
	data.Policies = make([]managedObjectStoragePolicyModel, len(policies))
	for i, policy := range policies {
		data.Policies[i].ServiceUUID = data.ServiceUUID
		data.Policies[i].ARN = types.StringPointerValue(policy.Arn)
		data.Policies[i].DefaultVersionID = types.StringPointerValue(policy.DefaultVersionId)
		data.Policies[i].Description = types.StringPointerValue(policy.Description)
		data.Policies[i].Document = types.StringPointerValue(policy.Document)
		data.Policies[i].Name = types.StringPointerValue(policy.Name)
		data.Policies[i].System = types.BoolPointerValue(policy.System)
		if policy.AttachmentCount != nil {
			data.Policies[i].AttachmentCount = types.Int64Value(int64(*policy.AttachmentCount))
		}
		if policy.CreatedAt != nil {
			data.Policies[i].CreatedAt = types.StringValue(policy.CreatedAt.String())
		}
		if policy.UpdatedAt != nil {
			data.Policies[i].UpdatedAt = types.StringValue(policy.UpdatedAt.String())
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
