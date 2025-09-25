package managedobjectstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &managedObjectStoragePolicyResource{}
	_ resource.ResourceWithConfigure   = &managedObjectStoragePolicyResource{}
	_ resource.ResourceWithImportState = &managedObjectStoragePolicyResource{}
)

func NewManagedObjectStoragePolicyResource() resource.Resource {
	return &managedObjectStoragePolicyResource{}
}

type managedObjectStoragePolicyResource struct {
	client *service.Service
}

func (r *managedObjectStoragePolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage_policy"
}

// Configure adds the provider configured client to the resource.
func (r *managedObjectStoragePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type policyModel struct {
	ARN              types.String `tfsdk:"arn"`
	AttachmentCount  types.Int64  `tfsdk:"attachment_count"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DefaultVersionID types.String `tfsdk:"default_version_id"`
	Description      types.String `tfsdk:"description"`
	Document         types.String `tfsdk:"document"`
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	ServiceUUID      types.String `tfsdk:"service_uuid"`
	System           types.Bool   `tfsdk:"system"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (r *managedObjectStoragePolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an UpCloud Managed Object Storage policy.",
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Description: "Policy ARN.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"attachment_count": schema.Int64Attribute{
				Description: "Attachment count.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation time.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_version_id": schema.StringAttribute{
				Description: "Default version id.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the policy.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"document": schema.StringAttribute{
				Description: "Policy document, URL-encoded compliant with RFC 3986. Extra whitespace and escapes are ignored when determining if the document has changed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						policyDocumentRequiresReplace,
						policyDocumentRequiresReplaceDescription,
						policyDocumentRequiresReplaceDescription,
					),
				},
			},
			"id": schema.StringAttribute{
				Description: "ID of the policy. ID is in {object storage UUID}/{policy name} format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Policy name.",
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
			"system": schema.BoolAttribute{
				Description: "Defines whether the policy was set up by the system.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Update time.",
				Computed:    true,
			},
		},
	}
}

func setPolicyValues(_ context.Context, data *policyModel, policy *upcloud.ManagedObjectStoragePolicy) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	data.ARN = types.StringValue(policy.ARN)
	data.AttachmentCount = types.Int64Value(int64(policy.AttachmentCount))
	data.CreatedAt = types.StringValue(policy.CreatedAt.String())
	data.DefaultVersionID = types.StringValue(policy.DefaultVersionID)
	data.Description = types.StringValue(policy.Description)
	data.Name = types.StringValue(policy.Name)
	data.System = types.BoolValue(policy.System)
	data.UpdatedAt = types.StringValue(policy.UpdatedAt.String())

	apiDocument, diags := normalizePolicyDocument(policy.Document)
	respDiagnostics.Append(diags...)
	// Document is required, so it should only be empty during import.
	if data.Document.IsNull() {
		data.Document = types.StringValue(apiDocument)
	}

	configDocument, diags := normalizePolicyDocument(data.Document.ValueString())
	respDiagnostics.Append(diags...)

	if configDocument != apiDocument {
		respDiagnostics.AddError(
			"Configured policy document does not match the policy document in the API response",
			fmt.Sprintf("Configured:   %s\nAPI response: %s", configDocument, apiDocument),
		)
	}

	return respDiagnostics
}

func (r *managedObjectStoragePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data policyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.Name.ValueString()))

	apiReq := &request.CreateManagedObjectStoragePolicyRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Document:    data.Document.ValueString(),
		ServiceUUID: data.ServiceUUID.ValueString(),
	}

	policy, err := r.client.CreateManagedObjectStoragePolicy(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage policy",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setPolicyValues(ctx, &data, policy)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStoragePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data policyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	var serviceUUID, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &serviceUUID, &name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal managed object storage policy ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	data.ServiceUUID = types.StringValue(serviceUUID)

	policy, err := r.client.GetManagedObjectStoragePolicy(ctx, &request.GetManagedObjectStoragePolicyRequest{
		Name:        name,
		ServiceUUID: serviceUUID,
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read managed object storage policy details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setPolicyValues(ctx, &data, policy)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStoragePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data policyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetManagedObjectStoragePolicy(ctx, &request.GetManagedObjectStoragePolicyRequest{
		Name:        data.Name.ValueString(),
		ServiceUUID: data.ServiceUUID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object storage policy details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setPolicyValues(ctx, &data, policy)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStoragePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data policyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var serviceUUID, name string
	if err := utils.UnmarshalID(data.ID.ValueString(), &serviceUUID, &name); err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal managed object storage policy ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if err := r.client.DeleteManagedObjectStoragePolicy(ctx, &request.DeleteManagedObjectStoragePolicyRequest{
		ServiceUUID: serviceUUID,
		Name:        name,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage policy",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *managedObjectStoragePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ stringplanmodifier.RequiresReplaceIfFunc = policyDocumentRequiresReplace

const policyDocumentRequiresReplaceDescription = "Policy document requires replace if the document in state does not match planned document after removing whitespace and unnecessary escapes."

func policyDocumentRequiresReplace(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	var plan, state policyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	planDocument, diags := normalizePolicyDocument(plan.Document.ValueString())
	resp.Diagnostics.Append(diags...)
	stateDocument, diags := normalizePolicyDocument(state.Document.ValueString())
	resp.Diagnostics.Append(diags...)

	resp.RequiresReplace = planDocument != stateDocument
}

func normalizePolicyDocument(document string) (string, diag.Diagnostics) {
	errToDiags := func(err error) (diags diag.Diagnostics) {
		diags.AddError(
			"Unable to normalize object storage policy document",
			utils.ErrorDiagnosticDetail(err),
		)
		return diags
	}

	unescaped, err := url.QueryUnescape(document)
	if err != nil {
		return "", errToDiags(err)
	}

	var unmarshaled interface{}
	err = json.Unmarshal([]byte(unescaped), &unmarshaled)
	if err != nil {
		return "", errToDiags(err)
	}

	marshaled, err := json.Marshal(unmarshaled)
	if err != nil {
		return "", errToDiags(err)
	}

	return url.QueryEscape(string(marshaled)), nil
}
