package loadbalancer

import (
	"context"
	"fmt"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
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
	client *service.Service
}

func (r *manualCertificateBundleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_manual_certificate_bundle"
}

// Configure adds the provider configured client to the resource.
func (r *manualCertificateBundleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type manualCertificateBundleModel struct {
	Certificate      types.String `tfsdk:"certificate"`
	ID               types.String `tfsdk:"id"`
	Intermediates    types.String `tfsdk:"intermediates"`
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

func setManualCertificateBundleValues(_ context.Context, data *manualCertificateBundleModel, bundle *upcloud.LoadBalancerCertificateBundle) diag.Diagnostics {
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

	apiReq := request.CreateLoadBalancerCertificateBundleRequest{
		Certificate:   data.Certificate.ValueString(),
		Intermediates: data.Intermediates.ValueString(),
		Name:          data.Name.ValueString(),
		PrivateKey:    data.PrivateKey.ValueString(),
		Type:          upcloud.LoadBalancerCertificateBundleTypeManual,
	}

	bundle, err := r.client.CreateLoadBalancerCertificateBundle(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create loadbalancer manual certificate bundle",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	data.ID = types.StringValue(bundle.UUID)

	resp.Diagnostics.Append(setManualCertificateBundleValues(ctx, &data, bundle)...)
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

	bundle, err := r.client.GetLoadBalancerCertificateBundle(ctx, &request.GetLoadBalancerCertificateBundleRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read loadbalancer manual certificate bundle details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setManualCertificateBundleValues(ctx, &data, bundle)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *manualCertificateBundleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data manualCertificateBundleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := &request.ModifyLoadBalancerCertificateBundleRequest{
		UUID:        data.ID.ValueString(),
		Name:        data.Name.ValueString(),
		Certificate: data.Certificate.ValueString(),
		// Use ValueString() to get empty string if not set, this will clear the intermediates
		Intermediates: upcloud.StringPtr(data.Intermediates.ValueString()),
		PrivateKey:    data.PrivateKey.ValueString(),
	}

	bundle, err := r.client.ModifyLoadBalancerCertificateBundle(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer manual certificate bundle",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setManualCertificateBundleValues(ctx, &data, bundle)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *manualCertificateBundleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data manualCertificateBundleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if err := r.client.DeleteLoadBalancerCertificateBundle(ctx, &request.DeleteLoadBalancerCertificateBundleRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer manual certificate bundle",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *manualCertificateBundleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
