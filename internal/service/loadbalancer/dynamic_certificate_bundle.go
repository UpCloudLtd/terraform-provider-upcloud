package loadbalancer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/google/uuid"

	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &dynamicCertificateBundleResource{}
	_ resource.ResourceWithConfigure   = &dynamicCertificateBundleResource{}
	_ resource.ResourceWithImportState = &dynamicCertificateBundleResource{}
)

func NewDynamicCertificateBundleResource() resource.Resource {
	return &dynamicCertificateBundleResource{}
}

type dynamicCertificateBundleResource struct {
	client *v9.ClientWithResponses
}

func (r *dynamicCertificateBundleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_dynamic_certificate_bundle"
}

// Configure adds the provider configured client to the resource.
func (r *dynamicCertificateBundleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

type dynamicCertificateBundleModel struct {
	certificateBundleCommonModel

	Hostnames types.List   `tfsdk:"hostnames"`
	KeyType   types.String `tfsdk:"key_type"`
}

func (r *dynamicCertificateBundleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource represents dynamic certificate bundle",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the certificate bundle.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostnames": schema.ListAttribute{
				MarkdownDescription: "Certificate hostnames.",
				ElementType:         types.StringType,
				Required:            true,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 100),
				},
			},
			"key_type": schema.StringAttribute{
				MarkdownDescription: "Private key type (`rsa` / `ecdsa`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("rsa", "ecdsa"),
				},
			},
			"labels": utils.LabelsAttribute("dynamic certificate bundle"),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the certificate bundle. Must be unique within customer account.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(validNameRegexp, validNameMessage),
				},
			},
			"not_after": schema.StringAttribute{
				MarkdownDescription: "The time after which a certificate is no longer valid.",
				Computed:            true,
			},
			"not_before": schema.StringAttribute{
				MarkdownDescription: "The time on which a certificate becomes valid.",
				Computed:            true,
			},
			"operational_state": schema.StringAttribute{
				MarkdownDescription: "The service operational state indicates the service's current operational, effective state. Managed by the system.",
				Computed:            true,
			},
		},
	}
}

func setDynamicCertificateBundleValues(ctx context.Context, data *dynamicCertificateBundleModel, bundle *v9.CreateLoadBalancerCertificateBundle201) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	respDiagnostics.Append(setCertificateBundleCommonValues(ctx, &data.certificateBundleCommonModel, bundle)...)

	data.Hostnames, diags = types.ListValueFrom(ctx, data.Hostnames.ElementType(ctx), bundle.Hostnames)
	respDiagnostics.Append(diags...)

	data.KeyType = types.StringPointerValue((*string)(bundle.KeyType))

	return respDiagnostics
}

func (r *dynamicCertificateBundleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data dynamicCertificateBundleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var hostnames []string
	if !data.Hostnames.IsNull() && !data.Hostnames.IsUnknown() {
		resp.Diagnostics.Append(data.Hostnames.ElementsAs(ctx, &hostnames, false)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var labelsMap map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelsMap, false)...)
	}
	labels := labelsMapToV9Slice(labelsMap)

	keyType := utils.ValueStringOrNil(data.KeyType)

	apiReq := v9.CreateLoadBalancerCertificateBundleJSONRequestBody{
		Type:      v9.LoadBalancerCertificateBundleCreateTypeDynamic,
		Name:      data.Name.ValueString(),
		KeyType:   (*v9.LoadBalancerCertificateBundleCreateKeyType)(keyType),
		Hostnames: &hostnames,
		Labels:    &labels,
	}

	apiResp, err := r.client.CreateLoadBalancerCertificateBundleWithResponse(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create loadbalancer dynamic certificate bundle",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}
	if apiResp.StatusCode() != http.StatusCreated || apiResp.JSON201 == nil {
		resp.Diagnostics.AddError(
			"Unable to create loadbalancer dynamic certificate bundle",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}

	data.ID = types.StringValue(apiResp.JSON201.Uuid.String())

	resp.Diagnostics.Append(setDynamicCertificateBundleValues(ctx, &data, apiResp.JSON201)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dynamicCertificateBundleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dynamicCertificateBundleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	serviceUUID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	apiResp, err := r.client.GetLoadBalancerCertificateBundleWithResponse(ctx, serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read loadbalancer dynamic certificate bundle details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if apiResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if apiResp.StatusCode() != http.StatusOK || apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to read loadbalancer dynamic certificate bundle details",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}

	resp.Diagnostics.Append(setDynamicCertificateBundleValues(ctx, &data, apiResp.JSON200)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dynamicCertificateBundleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dynamicCertificateBundleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var hostnames []string
	if !data.Hostnames.IsNull() && !data.Hostnames.IsUnknown() {
		resp.Diagnostics.Append(data.Hostnames.ElementsAs(ctx, &hostnames, false)...)
	}

	var labelsMap map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelsMap, false)...)
	}
	labels := labelsMapToV9Slice(labelsMap)

	modify := v9.LoadBalancerCertificateBundleDynamicModify{
		Name:      utils.ValueStringOrNil(data.Name),
		Hostnames: &hostnames,
		Labels:    &labels,
	}

	var apiReq v9.ModifyLoadBalancerCertificateBundleJSONRequestBody
	err := apiReq.FromLoadBalancerCertificateBundleDynamicModify(modify)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create API request for loadbalancer dynamic certificate bundle modification",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	serviceUUID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	apiResp, err := r.client.ModifyLoadBalancerCertificateBundleWithResponse(ctx, serviceUUID, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer dynamic certificate bundle",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if apiResp.StatusCode() != http.StatusOK || apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer dynamic certificate bundle",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}

	resp.Diagnostics.Append(setDynamicCertificateBundleValues(ctx, &data, apiResp.JSON200)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dynamicCertificateBundleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dynamicCertificateBundleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	serviceUUID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	apiResp, err := r.client.DeleteLoadBalancerCertificateBundleWithResponse(ctx, serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer dynamic certificate bundle",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if apiResp.StatusCode() != http.StatusNoContent && apiResp.StatusCode() != http.StatusAccepted && apiResp.StatusCode() != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer dynamic certificate bundle",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}
}

func (r *dynamicCertificateBundleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
