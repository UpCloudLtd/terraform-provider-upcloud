package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"

	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
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
	_ resource.Resource                 = &manualCertificateBundleResource{}
	_ resource.ResourceWithConfigure    = &manualCertificateBundleResource{}
	_ resource.ResourceWithImportState  = &manualCertificateBundleResource{}
	_ resource.ResourceWithUpgradeState = &manualCertificateBundleResource{}
)

func NewManualCertificateBundleResource() resource.Resource {
	return &manualCertificateBundleResource{}
}

type manualCertificateBundleResource struct {
	client *v9.ClientWithResponses
}

func (r *manualCertificateBundleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_manual_certificate_bundle"
}

// Configure adds the provider configured client to the resource.
func (r *manualCertificateBundleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

type manualCertificateBundleModel struct {
	Certificate      types.String `tfsdk:"certificate"`
	ID               types.String `tfsdk:"id"`
	Intermediates    types.String `tfsdk:"intermediates"`
	Labels           types.Map    `tfsdk:"labels"`
	Name             types.String `tfsdk:"name"`
	NotAfter         types.String `tfsdk:"not_after"`
	NotBefore        types.String `tfsdk:"not_before"`
	OperationalState types.String `tfsdk:"operational_state"`
	PrivateKey       types.String `tfsdk:"private_key"`
}

func (r *manualCertificateBundleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = manualCertificateBundleSchemaV1()
}

func manualCertificateBundleSchemaV1() schema.Schema {
	s := manualCertificateBundleSchemaV0()
	s.Version = 1
	intermediates, ok := s.Attributes["intermediates"].(schema.StringAttribute)
	if ok {
		intermediates.Default = stringdefault.StaticString("")
		s.Attributes["intermediates"] = intermediates
	}

	return s
}

func manualCertificateBundleSchemaV0() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "This resource represents manual certificate bundle",
		Attributes: map[string]schema.Attribute{
			"certificate": schema.StringAttribute{
				MarkdownDescription: "Certificate as base64 encoded string. Must be in PEM format.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the certificate bundle.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"intermediates": schema.StringAttribute{
				MarkdownDescription: "Intermediate certificates as base64 encoded string. Must be in PEM format.",
				Optional:            true,
				Computed:            true,
			},
			"labels": utils.LabelsAttribute("manual certificate bundle"),
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
			"private_key": schema.StringAttribute{
				MarkdownDescription: "Private key as base64 encoded string. Must be in PEM format.",
				Required:            true,
				Sensitive:           true,
			},
		},
		Version: 0,
	}
}

func (r *manualCertificateBundleResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := manualCertificateBundleSchemaV0()
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 to 1
		// No need to change the state as the upgrade is only adding a default value to `intermediates`
		0: {
			PriorSchema: &schemaV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData manualCertificateBundleModel
				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)
				if resp.Diagnostics.HasError() {
					return
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, priorStateData)...)
			},
		},
	}
}

func setManualCertificateBundleValues(_ context.Context, data *manualCertificateBundleModel, bundle *v9.CreateLoadBalancerCertificateBundle201) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	isImport := data.Certificate.IsNull()

	data.Name = types.StringValue(bundle.Name)
	data.NotAfter = types.StringValue(bundle.NotAfter.Format(time.RFC3339))
	data.NotBefore = types.StringValue(bundle.NotBefore.Format(time.RFC3339))
	data.OperationalState = types.StringValue(string(bundle.OperationalState))

	apiCertificate, diags := normalizeCertificate(bundle.Certificate)
	respDiagnostics.Append(diags...)
	if isImport {
		data.Certificate = types.StringValue(apiCertificate)
	} else {
		configCertificate, configDiags := normalizeCertificate(data.Certificate.ValueString())
		respDiagnostics.Append(configDiags...)

		if configCertificate != apiCertificate {
			respDiagnostics.AddError(
				"Configured certificate does not match the certificate in the API response",
				fmt.Sprintf("Configured:   %s\nAPI response: %s", configCertificate, apiCertificate),
			)
		}
	}

	apiIntermediates, diags := normalizeCertificate(bundle.Intermediates)
	respDiagnostics.Append(diags...)

	if isImport {
		data.Intermediates = types.StringValue(apiIntermediates)
	} else {
		configIntermediates, configDiags := normalizeCertificate(data.Intermediates.ValueString())
		respDiagnostics.Append(configDiags...)

		if configIntermediates != apiIntermediates {
			respDiagnostics.AddError(
				"Configured intermediates does not match the intermediates in the API response",
				fmt.Sprintf("Configured:   %s\nAPI response: %s", configIntermediates, apiIntermediates),
			)
		}
	}

	return respDiagnostics
}

func (r *manualCertificateBundleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data manualCertificateBundleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labelsMap map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelsMap, false)...)
	}
	labels := labelsMapToV9Slice(labelsMap)

	apiReq := v9.CreateLoadBalancerCertificateBundleJSONRequestBody{
		Certificate:   utils.ValueStringOrNil(data.Certificate),
		Intermediates: utils.ValueStringOrNil(data.Intermediates),
		Name:          data.Name.ValueString(),
		PrivateKey:    utils.ValueStringOrNil(data.PrivateKey),
		Type:          v9.Manual,
		Labels:        &labels,
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

	resp.Diagnostics.Append(setManualCertificateBundleValues(ctx, &data, apiResp.JSON201)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *manualCertificateBundleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data manualCertificateBundleModel
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
			"Unable to read loadbalancer manual certificate bundle details",
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
			"Unable to read loadbalancer manual certificate bundle details",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}

	resp.Diagnostics.Append(setManualCertificateBundleValues(ctx, &data, apiResp.JSON200)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *manualCertificateBundleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data manualCertificateBundleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var labelsMap map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelsMap, false)...)
	}
	labels := labelsMapToV9Slice(labelsMap)

	modify := v9.LoadBalancerCertificateBundleManualModify{
		Certificate:   utils.ValueStringOrNil(data.Certificate),
		Intermediates: utils.ValueStringOrNil(data.Intermediates),
		Labels:        &labels,
		Name:          utils.ValueStringOrNil(data.Name),
		PrivateKey:    utils.ValueStringOrNil(data.PrivateKey),
	}

	var apiReq v9.ModifyLoadBalancerCertificateBundleJSONRequestBody
	err := apiReq.FromLoadBalancerCertificateBundleManualModify(modify)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create API request for loadbalancer manual certificate bundle modification",
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
			"Unable to modify loadbalancer manual certificate bundle",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if apiResp.StatusCode() != http.StatusOK || apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer manual certificate bundle",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}

	resp.Diagnostics.Append(setManualCertificateBundleValues(ctx, &data, apiResp.JSON200)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *manualCertificateBundleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data manualCertificateBundleModel
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
			"Unable to delete loadbalancer manual certificate bundle",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if apiResp.StatusCode() != http.StatusNoContent && apiResp.StatusCode() != http.StatusAccepted && apiResp.StatusCode() != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer manual certificate bundle",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}
}

func (r *manualCertificateBundleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
