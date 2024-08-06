package loadbalancer

import (
	"context"
	"fmt"

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
	_ resource.Resource                = &frontendTLSConfigResource{}
	_ resource.ResourceWithConfigure   = &frontendTLSConfigResource{}
	_ resource.ResourceWithImportState = &frontendTLSConfigResource{}
)

func NewFrontendTLSConfigResource() resource.Resource {
	return &frontendTLSConfigResource{}
}

type frontendTLSConfigResource struct {
	client *service.Service
}

func (r *frontendTLSConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_frontend_tls_config"
}

// Configure adds the provider configured client to the resource.
func (r *frontendTLSConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type frontendTLSConfigModel struct {
	Frontend          types.String `tfsdk:"frontend"`
	CertificateBundle types.String `tfsdk:"certificate_bundle"`
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
}

func (r *frontendTLSConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource represents frontend TLS config",
		Attributes: map[string]schema.Attribute{
			"frontend": schema.StringAttribute{
				MarkdownDescription: "ID of the load balancer frontend to which the TLS config is connected.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate_bundle": schema.StringAttribute{
				MarkdownDescription: "Reference to certificate bundle ID.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the TLS config.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the TLS config. Must be unique within customer account.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(validNameRegexp, validNameMessage),
				},
			},
		},
	}
}

func setFrontendTLSConfigValues(_ context.Context, data *frontendTLSConfigModel, tlsConfig *upcloud.LoadBalancerFrontendTLSConfig) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	data.Name = types.StringValue(tlsConfig.Name)
	data.CertificateBundle = types.StringValue(tlsConfig.CertificateBundleUUID)

	return respDiagnostics
}

func (r *frontendTLSConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data frontendTLSConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var loadbalancer, frontendName string
	err := utils.UnmarshalID(data.Frontend.ValueString(), &loadbalancer, &frontendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer frontend name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	apiReq := request.CreateLoadBalancerFrontendTLSConfigRequest{
		ServiceUUID:  loadbalancer,
		FrontendName: frontendName,
		Config: request.LoadBalancerFrontendTLSConfig{
			Name:                  data.Name.ValueString(),
			CertificateBundleUUID: data.CertificateBundle.ValueString(),
		},
	}

	tlsConfig, err := r.client.CreateLoadBalancerFrontendTLSConfig(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create loadbalancer frontend TLS config",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	data.Frontend = types.StringValue(utils.MarshalID(loadbalancer, frontendName))
	data.ID = types.StringValue(utils.MarshalID(loadbalancer, frontendName, tlsConfig.Name))

	resp.Diagnostics.Append(setFrontendTLSConfigValues(ctx, &data, tlsConfig)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *frontendTLSConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data frontendTLSConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	var loadbalancer, frontendName string
	err := utils.UnmarshalID(data.Frontend.ValueString(), &loadbalancer, &frontendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer frontend name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	tlsConfig, err := r.client.GetLoadBalancerFrontendTLSConfig(ctx, &request.GetLoadBalancerFrontendTLSConfigRequest{
		ServiceUUID:  loadbalancer,
		FrontendName: frontendName,
		Name:         data.Name.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read loadbalancer frontend TLS config details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setFrontendTLSConfigValues(ctx, &data, tlsConfig)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *frontendTLSConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data frontendTLSConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var loadbalancer, frontendName string
	err := utils.UnmarshalID(data.Frontend.ValueString(), &loadbalancer, &frontendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer frontend name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	apiReq := &request.ModifyLoadBalancerFrontendTLSConfigRequest{
		ServiceUUID:  loadbalancer,
		FrontendName: frontendName,
		Name:         data.Name.ValueString(),
		Config: request.LoadBalancerFrontendTLSConfig{
			Name:                  data.Name.ValueString(),
			CertificateBundleUUID: data.CertificateBundle.ValueString(),
		},
	}

	tlsConfig, err := r.client.ModifyLoadBalancerFrontendTLSConfig(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer frontend TLS config",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setFrontendTLSConfigValues(ctx, &data, tlsConfig)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *frontendTLSConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data frontendTLSConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var loadbalancer, frontendName string
	err := utils.UnmarshalID(data.Frontend.ValueString(), &loadbalancer, &frontendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer frontend name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	if err := r.client.DeleteLoadBalancerFrontendTLSConfig(ctx, &request.DeleteLoadBalancerFrontendTLSConfigRequest{
		ServiceUUID:  loadbalancer,
		FrontendName: frontendName,
		Name:         data.Name.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer frontend TLS config",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *frontendTLSConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var loadbalancer, frontendName, name string
	err := utils.UnmarshalID(req.ID, &loadbalancer, &frontendName, &name)

	if err != nil || loadbalancer == "" || frontendName == "" || name == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: loadbalancer/frontend_name/name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("frontend"), utils.MarshalID(loadbalancer, frontendName))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}
