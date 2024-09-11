package loadbalancer

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &backendResource{}
	_ resource.ResourceWithConfigure   = &backendResource{}
	_ resource.ResourceWithImportState = &backendResource{}
)

func NewBackendResource() resource.Resource {
	return &backendResource{}
}

type backendResource struct {
	client *service.Service
}

func (r *backendResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_backend"
}

// Configure adds the provider configured client to the resource.
func (r *backendResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type backendModel struct {
	ID           types.String `tfsdk:"id"`
	LoadBalancer types.String `tfsdk:"loadbalancer"`
	Members      types.List   `tfsdk:"members"`
	Name         types.String `tfsdk:"name"`
	Properties   types.List   `tfsdk:"properties"`
	ResolverName types.String `tfsdk:"resolver_name"`
	TLSConfigs   types.List   `tfsdk:"tls_configs"`
}

type backendPropertiesModel struct {
	HealthCheckExpectedStatus types.Int64  `tfsdk:"health_check_expected_status"`
	HealthCheckFall           types.Int64  `tfsdk:"health_check_fall"`
	HealthCheckInterval       types.Int64  `tfsdk:"health_check_interval"`
	HealthCheckRise           types.Int64  `tfsdk:"health_check_rise"`
	HealthCheckTLSVerify      types.Bool   `tfsdk:"health_check_tls_verify"`
	HealthCheckType           types.String `tfsdk:"health_check_type"`
	HealthCheckURL            types.String `tfsdk:"health_check_url"`
	HTTP2Enabled              types.Bool   `tfsdk:"http2_enabled"`
	OutboundProxyProtocol     types.String `tfsdk:"outbound_proxy_protocol"`
	StickySessionCookieName   types.String `tfsdk:"sticky_session_cookie_name"`
	TimeoutServer             types.Int64  `tfsdk:"timeout_server"`
	TimeoutTunnel             types.Int64  `tfsdk:"timeout_tunnel"`
	TLSEnabled                types.Bool   `tfsdk:"tls_enabled"`
	TLSUseSystemCA            types.Bool   `tfsdk:"tls_use_system_ca"`
	TLSVerify                 types.Bool   `tfsdk:"tls_verify"`
}

func (r *backendResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents load balancer backend service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the backend. ID is in `{load balancer UUID}/{backend name}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"loadbalancer": schema.StringAttribute{
				MarkdownDescription: "UUID of the load balancer to which the backend is connected.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"members": schema.ListAttribute{
				Description: "Backend member server UUIDs. Members receive traffic dispatched from the frontends.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the backend. Must be unique within the load balancer service.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					nameValidator,
				},
			},
			"resolver_name": schema.StringAttribute{
				MarkdownDescription: "Domain name resolver used with dynamic type members.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"tls_configs": schema.ListAttribute{
				Description: "Set of TLS config names.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"properties": schema.ListNestedBlock{
				MarkdownDescription: "Backend properties. Properties can be set back to defaults by defining an empty `properties {}` block. For `terraform import`, an empty or non-empty block is also required.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"health_check_expected_status": schema.Int64Attribute{
							MarkdownDescription: "Expected HTTP status code returned by the customer application to mark server as healthy. Ignored for `tcp` `health_check_type`.",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(200),
							Validators: []validator.Int64{
								int64validator.Between(100, 599),
							},
						},
						"health_check_fall": schema.Int64Attribute{
							MarkdownDescription: "Sets how many failed health checks are allowed until the backend member is taken off from the rotation.",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(3),
							Validators: []validator.Int64{
								int64validator.Between(1, 100),
							},
						},
						"health_check_interval": schema.Int64Attribute{
							MarkdownDescription: "Interval between health checks in seconds.",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(10),
							Validators: []validator.Int64{
								int64validator.Between(1, 86400),
							},
						},
						"health_check_rise": schema.Int64Attribute{
							MarkdownDescription: "Sets how many successful health checks are required to put the backend member back into rotation.",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(3),
							Validators: []validator.Int64{
								int64validator.Between(1, 100),
							},
						},
						"health_check_tls_verify": schema.BoolAttribute{
							MarkdownDescription: "Enables certificate verification with the system CA certificate bundle. Works with https scheme in health_check_url, otherwise ignored.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"health_check_type": schema.StringAttribute{
							MarkdownDescription: "Health check type.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(string(upcloud.LoadBalancerHealthCheckTypeTCP)),
							Validators: []validator.String{
								stringvalidator.OneOf(string(upcloud.LoadBalancerHealthCheckTypeTCP), string(upcloud.LoadBalancerHealthCheckTypeHTTP)),
							},
						},
						"health_check_url": schema.StringAttribute{
							MarkdownDescription: "Target path for health check HTTP GET requests. Ignored for `tcp` `health_check_type`.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString("/"),
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 255),
							},
						},
						"http2_enabled": schema.BoolAttribute{
							MarkdownDescription: "Allow HTTP/2 connections to backend members by utilizing ALPN extension of TLS protocol, therefore it can only be enabled when tls_enabled is set to true. Note: members should support HTTP/2 for this setting to work.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"outbound_proxy_protocol": schema.StringAttribute{
							MarkdownDescription: "Enable outbound proxy protocol by setting the desired version. Defaults to empty string. Empty string disables proxy protocol.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
							Validators: []validator.String{
								stringvalidator.OneOf(
									"",
									string(upcloud.LoadBalancerProxyProtocolVersion1),
									string(upcloud.LoadBalancerProxyProtocolVersion2),
								),
							},
						},
						"sticky_session_cookie_name": schema.StringAttribute{
							MarkdownDescription: "Sets sticky session cookie name. Empty string disables sticky session.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
							Validators: []validator.String{
								stringvalidator.LengthBetween(0, 64),
							},
						},
						"timeout_server": schema.Int64Attribute{
							MarkdownDescription: "Backend server timeout in seconds.",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(10),
							Validators: []validator.Int64{
								int64validator.Between(1, 86400),
							},
						},
						"timeout_tunnel": schema.Int64Attribute{
							MarkdownDescription: "Maximum inactivity time on the client and server side for tunnels in seconds.",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(3600),
							Validators: []validator.Int64{
								int64validator.Between(1, 3024000),
							},
						},
						"tls_enabled": schema.BoolAttribute{
							MarkdownDescription: "Enables TLS connection from the load balancer to backend servers.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"tls_use_system_ca": schema.BoolAttribute{
							MarkdownDescription: "If enabled, then the system CA certificate bundle will be used for the certificate verification.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"tls_verify": schema.BoolAttribute{
							MarkdownDescription: "Enables backend servers certificate verification. Please make sure that TLS config with the certificate bundle of type authority attached to the backend or `tls_use_system_ca` enabled. Note: `tls_verify` has preference over `health_check_tls_verify` when `tls_enabled` in true.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 1),
				},
			},
		},
		Version: 1,
	}
}

func setBackendValues(ctx context.Context, data *backendModel, backend *upcloud.LoadBalancerBackend) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	isImport := data.LoadBalancer.ValueString() == ""

	var loadBalancer, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &loadBalancer, &name)
	if err != nil {
		diags.AddError(
			"Unable to unmarshal loadbalancer backend ID",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	data.LoadBalancer = types.StringValue(loadBalancer)
	data.Name = types.StringValue(data.Name.ValueString())

	members := make([]string, 0)
	for _, m := range backend.Members {
		members = append(members, m.Name)
	}

	data.Members, diags = types.ListValueFrom(ctx, types.StringType, members)
	respDiagnostics.Append(diags...)

	data.Name = types.StringValue(backend.Name)

	if !data.Properties.IsNull() || isImport {
		properties := make([]backendPropertiesModel, 1)
		properties[0].HealthCheckExpectedStatus = types.Int64Value(int64(backend.Properties.HealthCheckExpectedStatus))
		properties[0].HealthCheckFall = types.Int64Value(int64(backend.Properties.HealthCheckFall))
		properties[0].HealthCheckInterval = types.Int64Value(int64(backend.Properties.HealthCheckInterval))
		properties[0].HealthCheckRise = types.Int64Value(int64(backend.Properties.HealthCheckRise))
		properties[0].HealthCheckTLSVerify = types.BoolValue(*backend.Properties.HealthCheckTLSVerify)
		properties[0].HealthCheckType = types.StringValue(string(backend.Properties.HealthCheckType))
		properties[0].HealthCheckURL = types.StringValue(backend.Properties.HealthCheckURL)
		properties[0].HTTP2Enabled = types.BoolValue(*backend.Properties.HTTP2Enabled)
		properties[0].OutboundProxyProtocol = types.StringValue(string(backend.Properties.OutboundProxyProtocol))
		properties[0].StickySessionCookieName = types.StringValue(backend.Properties.StickySessionCookieName)
		properties[0].TimeoutServer = types.Int64Value(int64(backend.Properties.TimeoutServer))
		properties[0].TimeoutTunnel = types.Int64Value(int64(backend.Properties.TimeoutTunnel))
		properties[0].TLSEnabled = types.BoolValue(*backend.Properties.TLSEnabled)
		properties[0].TLSUseSystemCA = types.BoolValue(*backend.Properties.TLSUseSystemCA)
		properties[0].TLSVerify = types.BoolValue(*backend.Properties.TLSVerify)

		data.Properties, diags = types.ListValueFrom(ctx, data.Properties.ElementType(ctx), properties)
		respDiagnostics.Append(diags...)
	}

	data.ResolverName = types.StringValue(backend.Resolver)

	tlsConfigs := make([]string, 0)
	for _, t := range backend.TLSConfigs {
		tlsConfigs = append(tlsConfigs, t.Name)
	}

	data.TLSConfigs, diags = types.ListValueFrom(ctx, types.StringType, tlsConfigs)
	respDiagnostics.Append(diags...)

	return respDiagnostics
}

func buildBackendProperties(ctx context.Context, dataProperties types.List) (*upcloud.LoadBalancerBackendProperties, diag.Diagnostics) {
	if dataProperties.IsNull() {
		return nil, nil
	}

	var planProperties []backendPropertiesModel
	diags := dataProperties.ElementsAs(ctx, &planProperties, false)
	if diags.HasError() {
		return nil, diags
	}

	if len(planProperties) != 1 {
		return nil, nil
	}

	properties := planProperties[0]
	return &upcloud.LoadBalancerBackendProperties{
		TimeoutServer:             int(properties.TimeoutServer.ValueInt64()),
		TimeoutTunnel:             int(properties.TimeoutTunnel.ValueInt64()),
		HealthCheckTLSVerify:      properties.HealthCheckTLSVerify.ValueBoolPointer(),
		HealthCheckType:           upcloud.LoadBalancerHealthCheckType(properties.HealthCheckType.ValueString()),
		HealthCheckInterval:       int(properties.HealthCheckInterval.ValueInt64()),
		HealthCheckFall:           int(properties.HealthCheckFall.ValueInt64()),
		HealthCheckRise:           int(properties.HealthCheckRise.ValueInt64()),
		HealthCheckURL:            properties.HealthCheckURL.ValueString(),
		HealthCheckExpectedStatus: int(properties.HealthCheckExpectedStatus.ValueInt64()),
		StickySessionCookieName:   properties.StickySessionCookieName.ValueString(),
		OutboundProxyProtocol:     upcloud.LoadBalancerProxyProtocolVersion(properties.OutboundProxyProtocol.ValueString()),
		TLSEnabled:                properties.TLSEnabled.ValueBoolPointer(),
		TLSVerify:                 properties.TLSVerify.ValueBoolPointer(),
		TLSUseSystemCA:            properties.TLSUseSystemCA.ValueBoolPointer(),
		HTTP2Enabled:              properties.HTTP2Enabled.ValueBoolPointer(),
	}, diags
}

func (r *backendResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data backendModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	properties, diags := buildBackendProperties(ctx, data.Properties)
	resp.Diagnostics.Append(diags...)

	data.ID = types.StringValue(utils.MarshalID(data.LoadBalancer.ValueString(), data.Name.ValueString()))

	apiReq := request.CreateLoadBalancerBackendRequest{
		ServiceUUID: data.LoadBalancer.ValueString(),
		Backend: request.LoadBalancerBackend{
			Name:       data.Name.ValueString(),
			Resolver:   data.ResolverName.ValueString(),
			Properties: properties,
			Members:    []request.LoadBalancerBackendMember{},
			TLSConfigs: []request.LoadBalancerBackendTLSConfig{},
		},
	}

	backend, err := r.client.CreateLoadBalancerBackend(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create loadbalancer backend",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setBackendValues(ctx, &data, backend)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *backendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data backendModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	var loadBalancer, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &loadBalancer, &name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer backend ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	backend, err := r.client.GetLoadBalancerBackend(ctx, &request.GetLoadBalancerBackendRequest{
		ServiceUUID: loadBalancer,
		Name:        name,
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read loadbalancer backend details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setBackendValues(ctx, &data, backend)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *backendResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data backendModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var loadBalancer, name string
	if err := utils.UnmarshalID(data.ID.ValueString(), &loadBalancer, &name); err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer backend ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	properties, diags := buildBackendProperties(ctx, data.Properties)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := request.ModifyLoadBalancerBackendRequest{
		ServiceUUID: loadBalancer,
		Name:        name,
		Backend: request.ModifyLoadBalancerBackend{
			Name:       data.Name.ValueString(),
			Resolver:   data.ResolverName.ValueStringPointer(),
			Properties: properties,
		},
	}

	network, err := r.client.ModifyLoadBalancerBackend(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer backend",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setBackendValues(ctx, &data, network)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *backendResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data backendModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if err := r.client.DeleteLoadBalancerBackend(ctx, &request.DeleteLoadBalancerBackendRequest{
		ServiceUUID: data.LoadBalancer.ValueString(),
		Name:        data.Name.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer backend",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *backendResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
