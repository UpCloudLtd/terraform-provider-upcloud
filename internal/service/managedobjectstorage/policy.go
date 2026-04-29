package managedobjectstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &managedObjectStoragePolicyResource{}
	_ resource.ResourceWithConfigure   = &managedObjectStoragePolicyResource{}
	_ resource.ResourceWithImportState = &managedObjectStoragePolicyResource{}
)

func NewPolicyResource() resource.Resource {
	return &managedObjectStoragePolicyResource{}
}

type managedObjectStoragePolicyResource struct {
	client *v9.ClientWithResponses
}

func (r *managedObjectStoragePolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage_policy"
}

// Configure adds the provider configured client to the resource.
func (r *managedObjectStoragePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

type policyModel struct {
	managedObjectStoragePolicyModel

	ID types.String `tfsdk:"id"`
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
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
				Description: "Description of the policy. This property is immutable after creation.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					UseStateAfterCreate(),
				},
			},
			"document": schema.StringAttribute{
				Description: "Policy document, URL-encoded compliant with RFC 3986. Extra whitespace and escapes are ignored when determining if the document has changed.",
				Required:    true,
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

func setPolicyValues(_ context.Context, data *policyModel, policy *v9.ObjectStorage2PolicyDetailResponse) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	data.ARN = types.StringPointerValue(policy.Arn)
	data.DefaultVersionID = types.StringPointerValue(policy.DefaultVersionId)
	data.Name = types.StringPointerValue(policy.Name)
	data.System = types.BoolPointerValue(policy.System)
	if policy.AttachmentCount != nil {
		data.AttachmentCount = types.Int64Value(int64(*policy.AttachmentCount))
	}
	if policy.CreatedAt != nil {
		data.CreatedAt = types.StringValue(policy.CreatedAt.String())
	}
	if policy.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(policy.UpdatedAt.String())
	}

	if data.Description.IsNull() || data.Description.IsUnknown() {
		if policy.Description == nil || *policy.Description == "" {
			data.Description = types.StringNull()
		} else {
			data.Description = types.StringPointerValue(policy.Description)
		}
	}

	var apiDocStr string
	if policy.Document != nil {
		apiDocStr = *policy.Document
	}
	apiDocument, diags := normalizePolicyDocument(apiDocStr)
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

	svcUUID, err := uuid.Parse(data.ServiceUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage policy",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	apiReq := v9.CreateObjectStoragePolicyJSONRequestBody{
		Name:     data.Name.ValueString(),
		Document: data.Document.ValueString(),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		apiReq.Description = &desc
	}

	apiResp, err := r.client.CreateObjectStoragePolicyWithResponse(ctx, svcUUID, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage policy",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.JSON201 == nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage policy",
			utils.ErrorDiagnosticDetail(fmt.Errorf("unexpected response: %s", apiResp.HTTPResponse.Status)),
		)
		return
	}

	resp.Diagnostics.Append(setPolicyValues(ctx, &data, apiResp.JSON201)...)
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
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ServiceUUID = types.StringValue(serviceUUID)

	svcUUID, err := uuid.Parse(serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object storage policy details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	apiResp, err := r.client.GetObjectStoragePolicyWithResponse(ctx, svcUUID, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object storage policy details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.JSON200 == nil {
		if apiResp.HTTPResponse != nil && apiResp.HTTPResponse.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read managed object storage policy details",
				utils.ErrorDiagnosticDetail(fmt.Errorf("unexpected response: %s", apiResp.HTTPResponse.Status)),
			)
		}
		return
	}

	resp.Diagnostics.Append(setPolicyValues(ctx, &data, apiResp.JSON200)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStoragePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state policyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	planDoc, diags := normalizePolicyDocument(plan.Document.ValueString())
	resp.Diagnostics.Append(diags...)
	stateDoc, diags := normalizePolicyDocument(state.Document.ValueString())
	resp.Diagnostics.Append(diags...)

	svcUUID, err := uuid.Parse(plan.ServiceUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update managed object storage policy",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if planDoc != stateDoc {
		apiResp, err := r.client.CreateObjectStoragePolicyVersionWithResponse(
			ctx,
			svcUUID,
			plan.Name.ValueString(),
			v9.CreateObjectStoragePolicyVersionJSONRequestBody{
				Document:  plan.Document.ValueString(),
				IsDefault: true,
			},
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to create managed object storage policy version",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
		if apiResp.JSON201 == nil {
			resp.Diagnostics.AddError(
				"Unable to create managed object storage policy version",
				utils.ErrorDiagnosticDetail(fmt.Errorf("unexpected response: %s", apiResp.HTTPResponse.Status)),
			)
			return
		}
	}

	readResp, err := r.client.GetObjectStoragePolicyWithResponse(ctx, svcUUID, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object storage policy details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if readResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object storage policy details",
			utils.ErrorDiagnosticDetail(fmt.Errorf("unexpected response: %s", readResp.HTTPResponse.Status)),
		)
		return
	}

	resp.Diagnostics.Append(setPolicyValues(ctx, &plan, readResp.JSON200)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *managedObjectStoragePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data policyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var serviceUUID, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svcUUID, err := uuid.Parse(serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage policy",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	apiResp, err := r.client.DeleteObjectStoragePolicyWithResponse(ctx, svcUUID, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage policy",
			utils.ErrorDiagnosticDetail(err),
		)
	} else if apiResp.StatusCode() < 200 || apiResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage policy",
			objectStorageAPIErrorDetail(apiResp.ApplicationproblemJSONDefault, apiResp.Body),
		)
	}
}

func (r *managedObjectStoragePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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

	var unmarshaled map[string]interface{}
	err = json.Unmarshal([]byte(unescaped), &unmarshaled)
	if err != nil {
		return "", errToDiags(err)
	}

	// Use list type for Action field, because API converts single string to list.
	if statements, ok := unmarshaled["Statement"].([]interface{}); ok {
		for _, statement := range statements {
			if statement, ok := statement.(map[string]interface{}); ok {
				if action, ok := statement["Action"].(string); ok {
					statement["Action"] = []string{action}
				}
			}
		}
	}

	marshaled, err := json.Marshal(unmarshaled)
	if err != nil {
		return "", errToDiags(err)
	}

	return url.QueryEscape(string(marshaled)), nil
}
