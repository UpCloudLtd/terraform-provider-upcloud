package loadbalancer

import (
	"context"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

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
	_ resource.Resource                = &manualCertificateBundleResource{}
	_ resource.ResourceWithConfigure   = &manualCertificateBundleResource{}
	_ resource.ResourceWithImportState = &manualCertificateBundleResource{}
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
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource represents manual certificate bundle",
		Attributes: map[string]schema.Attribute{
			"certificate": schema.StringAttribute{
				MarkdownDescription: "Certificate within base64 string must be in PEM format.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed ID of the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"intermediates": schema.StringAttribute{
				MarkdownDescription: "Intermediate certificates within base64 string must be in PEM format.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the bundle must be unique within customer account.",
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
				MarkdownDescription: "Private key within base64 string must be in PEM format.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func setManualCertificateBundleValues(_ context.Context, data *manualCertificateBundleModel, bundle *upcloud.LoadBalancerCertificateBundle) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	data.Certificate = types.StringValue(bundle.Certificate)
	data.ID = types.StringValue(bundle.UUID)
	data.Intermediates = types.StringValue(bundle.Intermediates)
	data.Name = types.StringValue(bundle.Name)
	data.NotAfter = types.StringValue(bundle.NotAfter.Format(time.RFC3339))
	data.NotBefore = types.StringValue(bundle.NotBefore.Format(time.RFC3339))
	data.OperationalState = types.StringValue(string(bundle.OperationalState))

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

	apiReq := &request.ModifyLoadBalancerCertificateBundleRequest{
		UUID:          data.ID.ValueString(),
		Name:          data.Name.ValueString(),
		Certificate:   data.Certificate.ValueString(),
		Intermediates: data.Intermediates.ValueStringPointer(),
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
