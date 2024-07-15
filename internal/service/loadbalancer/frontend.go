package loadbalancer

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &frontendResource{}
	_ resource.ResourceWithConfigure   = &frontendResource{}
	_ resource.ResourceWithImportState = &frontendResource{}
)

func NewFrontendResource() resource.Resource {
	return &frontendResource{}
}

type frontendResource struct {
	client *service.Service
}

func (r *frontendResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_frontend"
}

// Configure adds the provider configured client to the resource.
func (r *frontendResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type frontendModel struct {
	ID                 types.String `tfsdk:"id"`
	LoadBalancer       types.String `tfsdk:"loadbalancer"`
	Name               types.String `tfsdk:"name"`
	Mode               types.String `tfsdk:"mode"`
	Port               types.Int64  `tfsdk:"port"`
	DefaultBackendName types.String `tfsdk:"default_backend_name"`
	Rules              types.List   `tfsdk:"rules"`
	TLSConfigs         types.List   `tfsdk:"tls_configs"`
	Networks           types.Set    `tfsdk:"networks"`
	Properties         types.List   `tfsdk:"properties"`
}

type propertiesModel struct {
	TimeoutClient        types.Int64 `tfsdk:"timeout_client"`
	InboundProxyProtocol types.Bool  `tfsdk:"inbound_proxy_protocol"`
	HTTP2Enabled         types.Bool  `tfsdk:"http2_enabled"`
}

type networkModel struct {
	Name types.String `tfsdk:"name"`
}

func (r *frontendResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents load balancer frontend service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the frontend. ID is `{load balancer UUID}/{frontend name}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"loadbalancer": schema.StringAttribute{
				MarkdownDescription: "UUID of the load balancer to which the frontend is connected.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the frontend. Must be unique within the load balancer service.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					nameValidator,
				},
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "When load balancer operating in `tcp` mode it acts as a layer 4 proxy. In `http` mode it acts as a layer 7 proxy.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(upcloud.LoadBalancerModeHTTP), string(upcloud.LoadBalancerModeTCP)),
				},
			},
			"port": schema.Int64Attribute{
				Description: "Port to listen for incoming requests.",
				Required:    true,
				Validators: []validator.Int64{
					portValidator,
				},
			},
			"default_backend_name": schema.StringAttribute{
				MarkdownDescription: "The name of the default backend where traffic will be routed. Note, default backend can be overwritten in frontend rules.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rules": schema.ListAttribute{
				Description: "Set of frontend rule names.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"tls_configs": schema.ListAttribute{
				Description: "Set of TLS config names.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"networks": schema.SetNestedBlock{
				MarkdownDescription: "Networks that frontend will be listening. Networks are required if load balancer has `networks` defined. This field will be required when deprecated field `network` is removed from load balancer resource.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the load balancer network.",
							Required:    true,
						},
					},
				},
				Validators: []validator.Set{
					setvalidator.SizeBetween(0, 100),
				},
			},
			"properties": schema.ListNestedBlock{
				MarkdownDescription: "Frontend properties. Properties can set back to defaults by defining empty `properties {}` block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"timeout_client": schema.Int64Attribute{
							Description: "Client request timeout in seconds.",
							Computed:    true,
							Optional:    true,
							Default:     int64default.StaticInt64(10),
							Validators: []validator.Int64{
								int64validator.Between(1, 86400),
							},
						},
						"inbound_proxy_protocol": schema.BoolAttribute{
							Description: "Enable or disable inbound proxy protocol support.",
							Computed:    true,
							Optional:    true,
							Default:     booldefault.StaticBool(false),
						},
						"http2_enabled": schema.BoolAttribute{
							Description: "Enable or disable HTTP/2 support.",
							Computed:    true,
							Optional:    true,
							Default:     booldefault.StaticBool(false),
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 1),
				},
			},
		},
	}
}

func setValues(ctx context.Context, data *frontendModel, frontend *upcloud.LoadBalancerFrontend) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	isImport := data.LoadBalancer.ValueString() == ""

	data.Name = types.StringValue(frontend.Name)
	data.Mode = types.StringValue(string(frontend.Mode))
	data.Port = types.Int64Value(int64(frontend.Port))
	data.DefaultBackendName = types.StringValue(frontend.DefaultBackend)

	rules := make([]string, 0)
	for _, r := range frontend.Rules {
		rules = append(rules, r.Name)
	}

	data.Rules, diags = types.ListValueFrom(ctx, types.StringType, rules)
	respDiagnostics.Append(diags...)

	tlsConfigs := make([]string, 0)
	for _, t := range frontend.TLSConfigs {
		tlsConfigs = append(tlsConfigs, t.Name)
	}

	data.TLSConfigs, diags = types.ListValueFrom(ctx, types.StringType, tlsConfigs)
	respDiagnostics.Append(diags...)

	if !data.Properties.IsNull() || isImport {
		properties := make([]propertiesModel, 1)
		properties[0].TimeoutClient = types.Int64Value(int64(frontend.Properties.TimeoutClient))
		properties[0].InboundProxyProtocol = asBool(frontend.Properties.InboundProxyProtocol)
		properties[0].HTTP2Enabled = asBool(frontend.Properties.HTTP2Enabled)

		data.Properties, diags = types.ListValueFrom(ctx, data.Properties.ElementType(ctx), properties)
		respDiagnostics.Append(diags...)
	}

	if !data.Networks.IsNull() || isImport {
		networks := make([]networkModel, len(frontend.Networks))
		for i, net := range frontend.Networks {
			networks[i].Name = types.StringValue(net.Name)
		}

		data.Networks, diags = types.SetValueFrom(ctx, data.Networks.ElementType(ctx), networks)
		respDiagnostics.Append(diags...)
	}

	return respDiagnostics
}

func buildNetworks(ctx context.Context, dataNetworks types.Set) ([]upcloud.LoadBalancerFrontendNetwork, diag.Diagnostics) {
	var planNetworks []networkModel
	respDiagnostics := dataNetworks.ElementsAs(ctx, &planNetworks, false)

	networks := make([]upcloud.LoadBalancerFrontendNetwork, 0)

	for _, net := range planNetworks {
		networks = append(networks, upcloud.LoadBalancerFrontendNetwork{
			Name: net.Name.ValueString(),
		})
	}

	return networks, respDiagnostics
}

func buildProperties(ctx context.Context, dataProperties types.List) (*upcloud.LoadBalancerFrontendProperties, diag.Diagnostics) {
	if dataProperties.IsNull() {
		return nil, nil
	}

	var planProperties []propertiesModel
	diags := dataProperties.ElementsAs(ctx, &planProperties, false)
	if diags.HasError() {
		return nil, diags
	}

	if len(planProperties) != 1 {
		return nil, nil
	}

	properties := planProperties[0]
	return &upcloud.LoadBalancerFrontendProperties{
		TimeoutClient:        int(properties.TimeoutClient.ValueInt64()),
		InboundProxyProtocol: properties.InboundProxyProtocol.ValueBoolPointer(),
		HTTP2Enabled:         properties.HTTP2Enabled.ValueBoolPointer(),
	}, diags
}

func (r *frontendResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data frontendModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	networks, diags := buildNetworks(ctx, data.Networks)
	resp.Diagnostics.Append(diags...)

	properties, diags := buildProperties(ctx, data.Properties)
	resp.Diagnostics.Append(diags...)

	data.ID = types.StringValue(utils.MarshalID(data.LoadBalancer.ValueString(), data.Name.ValueString()))
	apiReq := request.CreateLoadBalancerFrontendRequest{
		ServiceUUID: data.LoadBalancer.ValueString(),
		Frontend: request.LoadBalancerFrontend{
			Name:           data.Name.ValueString(),
			Mode:           upcloud.LoadBalancerMode(data.Mode.ValueString()),
			Port:           int(data.Port.ValueInt64()),
			DefaultBackend: data.DefaultBackendName.ValueString(),
			Properties:     properties,
			Networks:       networks,
		},
	}

	frontend, err := r.client.CreateLoadBalancerFrontend(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create loadbalancer frontend",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, frontend)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *frontendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data frontendModel
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
			"Unable to unmarshal loadbalancer frontend ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	network, err := r.client.GetLoadBalancerFrontend(ctx, &request.GetLoadBalancerFrontendRequest{
		ServiceUUID: loadBalancer,
		Name:        name,
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read loadbalancer frontend details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, network)...)

	data.LoadBalancer = types.StringValue(loadBalancer)
	data.Name = types.StringValue(name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *frontendResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data frontendModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var loadBalancer, name string
	if err := utils.UnmarshalID(data.ID.ValueString(), &loadBalancer, &name); err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer frontend ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	networks, diags := buildNetworks(ctx, data.Networks)
	resp.Diagnostics.Append(diags...)

	properties, diags := buildProperties(ctx, data.Properties)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := request.ModifyLoadBalancerFrontendRequest{
		ServiceUUID: loadBalancer,
		Name:        name,
		Frontend: request.ModifyLoadBalancerFrontend{
			Name:           data.Name.ValueString(),
			Mode:           upcloud.LoadBalancerMode(data.Mode.ValueString()),
			Port:           int(data.Port.ValueInt64()),
			DefaultBackend: data.DefaultBackendName.ValueString(),
			Networks:       networks,
			Properties:     properties,
		},
	}

	network, err := r.client.ModifyLoadBalancerFrontend(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer frontend",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, network)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *frontendResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data frontendModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if err := r.client.DeleteLoadBalancerFrontend(ctx, &request.DeleteLoadBalancerFrontendRequest{
		ServiceUUID: data.LoadBalancer.ValueString(),
		Name:        data.Name.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer frontend",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *frontendResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
