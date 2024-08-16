package loadbalancer

import (
	"context"

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
	_ resource.Resource                = &backendTLSConfigResource{}
	_ resource.ResourceWithConfigure   = &backendTLSConfigResource{}
	_ resource.ResourceWithImportState = &backendTLSConfigResource{}
)

func NewBackendTLSConfigResource() resource.Resource {
	return &backendTLSConfigResource{}
}

type backendTLSConfigResource struct {
	client *service.Service
}

func (r *backendTLSConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_backend_tls_config"
}

// Configure adds the provider configured client to the resource.
func (r *backendTLSConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type backendTLSConfigModel struct {
	Backend           types.String `tfsdk:"backend"`
	CertificateBundle types.String `tfsdk:"certificate_bundle"`
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
}

func (r *backendTLSConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource represents backend TLS config",
		Attributes: map[string]schema.Attribute{
			"backend": schema.StringAttribute{
				MarkdownDescription: "ID of the load balancer backend to which the TLS config is connected.",
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
				MarkdownDescription: "ID of the TLS config. ID is in `{load balancer UUID}/{backend name}/{TLS config name}` format.",
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

func setBackendTLSConfigValues(_ context.Context, data *backendTLSConfigModel, tlsConfig *upcloud.LoadBalancerBackendTLSConfig) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	isImport := data.Backend.ValueString() == ""

	if isImport || !data.Backend.IsNull() {
		var loadbalancer, backendName, name string

		err := utils.UnmarshalID(data.ID.ValueString(), &loadbalancer, &backendName, &name)
		if err != nil {
			respDiagnostics.AddError(
				"Unable to unmarshal loadbalancer backend name",
				utils.ErrorDiagnosticDetail(err),
			)
		}

		data.Backend = types.StringValue(utils.MarshalID(loadbalancer, backendName))
	}

	data.Name = types.StringValue(tlsConfig.Name)
	data.CertificateBundle = types.StringValue(tlsConfig.CertificateBundleUUID)

	return respDiagnostics
}

func (r *backendTLSConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data backendTLSConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var loadbalancer, backendName string
	err := utils.UnmarshalID(data.Backend.ValueString(), &loadbalancer, &backendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer backend name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	apiReq := request.CreateLoadBalancerBackendTLSConfigRequest{
		ServiceUUID: loadbalancer,
		BackendName: backendName,
		Config: request.LoadBalancerBackendTLSConfig{
			Name:                  data.Name.ValueString(),
			CertificateBundleUUID: data.CertificateBundle.ValueString(),
		},
	}

	tlsConfig, err := r.client.CreateLoadBalancerBackendTLSConfig(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create loadbalancer backend TLS config",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	data.ID = types.StringValue(utils.MarshalID(loadbalancer, backendName, tlsConfig.Name))

	resp.Diagnostics.Append(setBackendTLSConfigValues(ctx, &data, tlsConfig)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *backendTLSConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data backendTLSConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	var loadbalancer, backendName, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &loadbalancer, &backendName, &name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer backend name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	tlsConfig, err := r.client.GetLoadBalancerBackendTLSConfig(ctx, &request.GetLoadBalancerBackendTLSConfigRequest{
		ServiceUUID: loadbalancer,
		BackendName: backendName,
		Name:        name,
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read loadbalancer backend TLS config details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setBackendTLSConfigValues(ctx, &data, tlsConfig)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *backendTLSConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data backendTLSConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var loadbalancer, backendName, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &loadbalancer, &backendName, &name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer backend name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	apiReq := &request.ModifyLoadBalancerBackendTLSConfigRequest{
		ServiceUUID: loadbalancer,
		BackendName: backendName,
		Name:        name,
		Config: request.LoadBalancerBackendTLSConfig{
			Name:                  data.Name.ValueString(),
			CertificateBundleUUID: data.CertificateBundle.ValueString(),
		},
	}

	tlsConfig, err := r.client.ModifyLoadBalancerBackendTLSConfig(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer backend TLS config",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(loadbalancer, backendName, tlsConfig.Name))

	resp.Diagnostics.Append(setBackendTLSConfigValues(ctx, &data, tlsConfig)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *backendTLSConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data backendTLSConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var loadbalancer, backendName string
	err := utils.UnmarshalID(data.Backend.ValueString(), &loadbalancer, &backendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer backend name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	if err := r.client.DeleteLoadBalancerBackendTLSConfig(ctx, &request.DeleteLoadBalancerBackendTLSConfigRequest{
		ServiceUUID: loadbalancer,
		BackendName: backendName,
		Name:        data.Name.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer backend TLS config",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *backendTLSConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
