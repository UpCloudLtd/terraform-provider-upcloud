package loadbalancer

import (
	"context"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
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
	client *service.Service
}

func (r *dynamicCertificateBundleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_dynamic_certificate_bundle"
}

// Configure adds the provider configured client to the resource.
func (r *dynamicCertificateBundleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type dynamicCertificateBundleModel struct {
	ID               types.String `tfsdk:"id"`
	Hostnames        types.List   `tfsdk:"hostnames"`
	KeyType          types.String `tfsdk:"key_type"`
	Name             types.String `tfsdk:"name"`
	NotAfter         types.String `tfsdk:"not_after"`
	NotBefore        types.String `tfsdk:"not_before"`
	OperationalState types.String `tfsdk:"operational_state"`
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

func setDynamicCertificateBundleValues(ctx context.Context, data *dynamicCertificateBundleModel, bundle *upcloud.LoadBalancerCertificateBundle) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	data.Hostnames, diags = types.ListValueFrom(ctx, data.Hostnames.ElementType(ctx), bundle.Hostnames)
	respDiagnostics.Append(diags...)

	data.KeyType = types.StringValue(bundle.KeyType)
	data.Name = types.StringValue(bundle.Name)
	data.NotAfter = types.StringValue(bundle.NotAfter.Format(time.RFC3339))
	data.NotBefore = types.StringValue(bundle.NotBefore.Format(time.RFC3339))
	data.OperationalState = types.StringValue(string(bundle.OperationalState))

	return respDiagnostics
}

func (r *dynamicCertificateBundleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data dynamicCertificateBundleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var hostnames []string
	if !data.Hostnames.IsNull() && !data.Hostnames.IsUnknown() {
		resp.Diagnostics.Append(data.Hostnames.ElementsAs(ctx, &hostnames, false)...)
	}

	apiReq := request.CreateLoadBalancerCertificateBundleRequest{
		Type:      upcloud.LoadBalancerCertificateBundleTypeDynamic,
		Name:      data.Name.ValueString(),
		KeyType:   data.KeyType.ValueString(),
		Hostnames: hostnames,
	}

	bundle, err := r.client.CreateLoadBalancerCertificateBundle(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create loadbalancer dynamic certificate bundle",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	data.ID = types.StringValue(bundle.UUID)

	resp.Diagnostics.Append(setDynamicCertificateBundleValues(ctx, &data, bundle)...)
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

	bundle, err := r.client.GetLoadBalancerCertificateBundle(ctx, &request.GetLoadBalancerCertificateBundleRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read loadbalancer dynamic certificate bundle details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setDynamicCertificateBundleValues(ctx, &data, bundle)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dynamicCertificateBundleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dynamicCertificateBundleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var hostnames []string
	if !data.Hostnames.IsNull() && !data.Hostnames.IsUnknown() {
		resp.Diagnostics.Append(data.Hostnames.ElementsAs(ctx, &hostnames, false)...)
	}

	apiReq := &request.ModifyLoadBalancerCertificateBundleRequest{
		UUID:      data.ID.ValueString(),
		Name:      data.Name.ValueString(),
		Hostnames: hostnames,
	}

	bundle, err := r.client.ModifyLoadBalancerCertificateBundle(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer dynamic certificate bundle",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setDynamicCertificateBundleValues(ctx, &data, bundle)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dynamicCertificateBundleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dynamicCertificateBundleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if err := r.client.DeleteLoadBalancerCertificateBundle(ctx, &request.DeleteLoadBalancerCertificateBundleRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer dynamic certificate bundle",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *dynamicCertificateBundleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
