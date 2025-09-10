package loadbalancer

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &loadBalancerResource{}
	_ resource.ResourceWithConfigure   = &loadBalancerResource{}
	_ resource.ResourceWithImportState = &loadBalancerResource{}
)

func NewLoadBalancerResource() resource.Resource {
	return &loadBalancerResource{}
}

type loadBalancerResource struct {
	client *service.Service
}

func (r *loadBalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer"
}

// Configure adds the provider configured client to the resource.
func (r *loadBalancerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type loadBalancerModel struct {
	Backends         types.List   `tfsdk:"backends"`
	ConfiguredStatus types.String `tfsdk:"configured_status"`
	DNSName          types.String `tfsdk:"dns_name"`
	Frontends        types.List   `tfsdk:"frontends"`
	ID               types.String `tfsdk:"id"`
	IPAddresses      types.Set    `tfsdk:"ip_addresses"`
	Labels           types.Map    `tfsdk:"labels"`
	MaintenanceDOW   types.String `tfsdk:"maintenance_dow"`
	MaintenanceTime  types.String `tfsdk:"maintenance_time"`
	Name             types.String `tfsdk:"name"`
	Network          types.String `tfsdk:"network"`
	Networks         types.List   `tfsdk:"networks"`
	Nodes            types.List   `tfsdk:"nodes"`
	OperationalState types.String `tfsdk:"operational_state"`
	Plan             types.String `tfsdk:"plan"`
	Resolvers        types.List   `tfsdk:"resolvers"`
	Zone             types.String `tfsdk:"zone"`
}

type loadbalancerIPAddressModel struct {
	NetworkName types.String `tfsdk:"network_name"`
	Address     types.String `tfsdk:"address"`
}

type loadbalancerNetworkModel struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Family  types.String `tfsdk:"family"`
	Network types.String `tfsdk:"network"`
	DNSName types.String `tfsdk:"dns_name"`
	ID      types.String `tfsdk:"id"`
}

type loadbalancerNodeModel struct {
	OperationalState types.String `tfsdk:"operational_state"`
	Networks         types.List   `tfsdk:"networks"`
}

type loadbalancerNodeNetworkModel struct {
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	IPAddresses types.List   `tfsdk:"ip_addresses"`
}

func (m loadbalancerNodeNetworkModel) ElementType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"type": types.StringType,
			"ip_addresses": types.ListType{
				ElemType: loadBalancerNodeNetworkIPAddressModel{}.ElementType(),
			},
		},
	}
}

type loadBalancerNodeNetworkIPAddressModel struct {
	Address types.String `tfsdk:"address"`
	Listen  types.Bool   `tfsdk:"listen"`
}

func (m loadBalancerNodeNetworkIPAddressModel) ElementType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"address": types.StringType,
			"listen":  types.BoolType,
		},
	}
}

func (r *loadBalancerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents [Managed Load Balancer](https://upcloud.com/products/managed-load-balancer) service.",
		Attributes: map[string]schema.Attribute{
			"backends": schema.ListAttribute{
				MarkdownDescription: "Backends are groups of customer servers whose traffic should be balanced.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"configured_status": schema.StringAttribute{
				MarkdownDescription: "The service configured status indicates the service's current intended status. Managed by the customer.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(upcloud.LoadBalancerConfiguredStatusStarted)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(upcloud.LoadBalancerConfiguredStatusStarted),
						string(upcloud.LoadBalancerConfiguredStatusStopped),
					),
				},
			},
			"dns_name": schema.StringAttribute{
				DeprecationMessage:  "Use 'networks' to get network DNS name",
				MarkdownDescription: "DNS name of the load balancer",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"frontends": schema.ListAttribute{
				MarkdownDescription: "Frontends receive the traffic before dispatching it to the backends.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the load balancer.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ip_addresses": schema.SetNestedAttribute{
				MarkdownDescription: "Floating IP addresses connected to the load balancer.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"network_name": schema.StringAttribute{
							MarkdownDescription: "Name of the network where to attach the IP address.",
							Required:            true,
						},
						"address": schema.StringAttribute{
							Description: "Floating IP address to attach to the load balancer.",
							Required:    true,
						},
					},
				},
			},
			"labels": utils.LabelsAttribute("load balancer"),
			"maintenance_dow": schema.StringAttribute{
				MarkdownDescription: "The day of the week on which maintenance will be performed. If not provided, we will randomly select a weekend day. Valid values `monday|tuesday|wednesday|thursday|friday|saturday|sunday`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"maintenance_time": schema.StringAttribute{
				MarkdownDescription: "The time at which the maintenance will begin in UTC. A 2-hour timeframe has been allocated for maintenance. During this period, the multi-node production plans will not experience any downtime, while the one-node plans will have a downtime of 1-2 minutes. If not provided, we will randomly select an off-peak time. Needs to be a valid time format in UTC HH:MM:SSZ, for example `20:01:01Z`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service. Must be unique within customer account.",
				Required:            true,
				Validators: []validator.String{
					nameValidator,
				},
			},
			"network": schema.StringAttribute{
				DeprecationMessage:  "Use 'networks' to define networks attached to load balancer",
				MarkdownDescription: "Private network UUID where traffic will be routed. Must reside in load balancer zone.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nodes": schema.ListNestedAttribute{
				MarkdownDescription: "Nodes are instances running load balancer service",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"operational_state": schema.StringAttribute{
							MarkdownDescription: "Node's operational state. Managed by the system.",
							Computed:            true,
						},
						"networks": schema.ListNestedAttribute{
							MarkdownDescription: "Networks attached to the node",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										MarkdownDescription: "The name of the network",
										Computed:            true,
									},
									"type": schema.StringAttribute{
										MarkdownDescription: "The type of the network",
										Computed:            true,
									},
									"ip_addresses": schema.ListNestedAttribute{
										MarkdownDescription: "IP addresses attached to the network",
										Computed:            true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"address": schema.StringAttribute{
													MarkdownDescription: "Node's IP address",
													Computed:            true,
												},
												"listen": schema.BoolAttribute{
													MarkdownDescription: "Whether the node listens to the traffic",
													Computed:            true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"operational_state": schema.StringAttribute{
				MarkdownDescription: "The service operational state indicates the service's current operational, effective state. Managed by the system.",
				Computed:            true,
			},
			"plan": schema.StringAttribute{
				MarkdownDescription: "Plan which the service will have. You can list available load balancer plans with `upctl loadbalancer plans`",
				Required:            true,
			},
			"resolvers": schema.ListAttribute{
				MarkdownDescription: "Domain Name Resolvers.",
				ElementType:         types.StringType,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.StringAttribute{
				Description: "Zone in which the service will be hosted, e.g. `fi-hel1`. You can list available zones with `upctl zone list`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"networks": schema.ListNestedBlock{
				MarkdownDescription: "Attached Networks from where traffic consumed and routed. Private networks must reside in loadbalancer zone.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"dns_name": schema.StringAttribute{
							MarkdownDescription: "DNS name of the load balancer network",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"family": schema.StringAttribute{
							Description: "Network family. Currently only `IPv4` is supported.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(upcloud.LoadBalancerAddressFamilyIPv4),
								),
							},
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier of the network.",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the network. Must be unique within the service.",
							Required:            true,
							Validators: []validator.String{
								nameValidator,
								stringvalidator.LengthBetween(0, 65),
							},
						},
						"network": schema.StringAttribute{
							MarkdownDescription: "Private network UUID. Required for private networks and must reside in loadbalancer zone. For public network the field should be omitted.",
							Optional:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the network. Only one public network can be attached and at least one private network must be attached.",
							Required:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(upcloud.LoadBalancerNetworkTypePrivate),
									string(upcloud.LoadBalancerNetworkTypePublic),
								),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 8),
					listvalidator.ExactlyOneOf(
						path.Expressions{
							path.MatchRoot("network"),
						}...,
					),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
					getNetworksPlanModifier(),
				},
			},
		},
	}
}

func (r *loadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data loadBalancerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	var labelsMap map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelsMap, false)...)
	}
	labels := utils.NilAsEmptyList(utils.LabelsMapToSlice(labelsMap))

	var networks []request.LoadBalancerNetwork
	if !data.Networks.IsNull() && !data.Networks.IsUnknown() {
		var networkModels []loadbalancerNetworkModel
		resp.Diagnostics.Append(data.Networks.ElementsAs(ctx, &networkModels, false)...)

		for _, n := range networkModels {
			network := request.LoadBalancerNetwork{
				Name:   n.Name.ValueString(),
				Type:   upcloud.LoadBalancerNetworkType(n.Type.ValueString()),
				Family: upcloud.LoadBalancerAddressFamily(n.Family.ValueString()),
				UUID:   n.Network.ValueString(),
			}
			networks = append(networks, network)
		}
	}

	var ipAddresses []request.LoadBalancerIPAddress
	if !data.IPAddresses.IsNull() && !data.IPAddresses.IsUnknown() {
		var ipAddressModels []loadbalancerIPAddressModel
		resp.Diagnostics.Append(data.IPAddresses.ElementsAs(ctx, &ipAddressModels, false)...)

		for _, ipAddressModel := range ipAddressModels {
			ip := request.LoadBalancerIPAddress{
				NetworkName: ipAddressModel.NetworkName.ValueString(),
				Address:     ipAddressModel.Address.ValueString(),
			}
			ipAddresses = append(ipAddresses, ip)
		}
	}

	apiReq := request.CreateLoadBalancerRequest{
		Name:             data.Name.ValueString(),
		Plan:             data.Plan.ValueString(),
		Zone:             data.Zone.ValueString(),
		NetworkUUID:      data.Network.ValueString(),
		Networks:         networks,
		IPAddresses:      ipAddresses,
		ConfiguredStatus: upcloud.LoadBalancerConfiguredStatus(data.ConfiguredStatus.ValueString()),
		Frontends:        []request.LoadBalancerFrontend{},
		Backends:         []request.LoadBalancerBackend{},
		Resolvers:        []request.LoadBalancerResolver{},
		Labels:           labels,
		MaintenanceDOW:   upcloud.LoadBalancerMaintenanceDOW(data.MaintenanceDOW.ValueString()),
		MaintenanceTime:  data.MaintenanceTime.ValueString(),
	}

	loadBalancer, err := r.client.CreateLoadBalancer(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create loadbalancer",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(loadBalancer.UUID)

	loadBalancer, diags := waitForRunningState(ctx, r.client, data, "creation")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setLoadBalancerValues(ctx, &data, loadBalancer)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func waitForRunningState(ctx context.Context, client *service.Service, data loadBalancerModel, action string) (lb *upcloud.LoadBalancer, diags diag.Diagnostics) {
	if data.ConfiguredStatus.ValueString() == string(upcloud.LoadBalancerConfiguredStatusStarted) {
		var err error
		lb, err = client.WaitForLoadBalancerOperationalState(ctx, &request.WaitForLoadBalancerOperationalStateRequest{
			UUID:         data.ID.ValueString(),
			DesiredState: upcloud.LoadBalancerOperationalStateRunning,
		})
		if err != nil {
			diags.AddError(
				"Loadbalancer did not reach running state after "+action,
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
	}
	return
}

func setLoadBalancerValues(ctx context.Context, data *loadBalancerModel, loadbalancer *upcloud.LoadBalancer) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	backendNames := []string{}
	for _, backend := range loadbalancer.Backends {
		backendNames = append(backendNames, backend.Name)
	}
	data.Backends, diags = types.ListValueFrom(ctx, types.StringType, backendNames)
	respDiagnostics.Append(diags...)

	data.ConfiguredStatus = types.StringValue(string(loadbalancer.ConfiguredStatus))
	data.DNSName = types.StringValue(loadbalancer.DNSName)

	frontendNames := []string{}
	for _, frontend := range loadbalancer.Frontends {
		frontendNames = append(frontendNames, frontend.Name)
	}
	data.Frontends, diags = types.ListValueFrom(ctx, types.StringType, frontendNames)
	respDiagnostics.Append(diags...)

	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(loadbalancer.Labels))
	respDiagnostics.Append(diags...)

	data.MaintenanceDOW = types.StringValue(string(loadbalancer.MaintenanceDOW))
	data.MaintenanceTime = types.StringValue(loadbalancer.MaintenanceTime)
	data.Name = types.StringValue(loadbalancer.Name)
	data.Network = types.StringValue(loadbalancer.NetworkUUID)

	networks := make([]loadbalancerNetworkModel, len(loadbalancer.Networks))
	for i, network := range loadbalancer.Networks {
		dataNetwork := loadbalancerNetworkModel{
			Name:    types.StringValue(network.Name),
			Type:    types.StringValue(string(network.Type)),
			Family:  types.StringValue(string(network.Family)),
			DNSName: types.StringValue(network.DNSName),
			ID:      types.StringValue(utils.MarshalID(loadbalancer.UUID, network.Name)),
		}
		if network.Type == upcloud.LoadBalancerNetworkTypePrivate {
			dataNetwork.Network = types.StringValue(network.UUID)
		}
		networks[i] = dataNetwork
	}

	data.Networks, diags = types.ListValueFrom(ctx, data.Networks.ElementType(ctx), networks)
	respDiagnostics.Append(diags...)

	if data.IPAddresses.IsNull() {
		data.IPAddresses = types.SetNull(data.IPAddresses.ElementType(ctx))
	} else {
		ipAddresses := make([]loadbalancerIPAddressModel, len(loadbalancer.IPAddresses))
		for i, ip := range loadbalancer.IPAddresses {
			dataIPAddress := loadbalancerIPAddressModel{
				NetworkName: types.StringValue(ip.NetworkName),
				Address:     types.StringValue(ip.Address),
			}
			ipAddresses[i] = dataIPAddress
		}

		data.IPAddresses, diags = types.SetValueFrom(ctx, data.IPAddresses.ElementType(ctx), ipAddresses)
		respDiagnostics.Append(diags...)
	}

	nodes := make([]loadbalancerNodeModel, len(loadbalancer.Nodes))
	for i, node := range loadbalancer.Nodes {
		dataNode := loadbalancerNodeModel{
			OperationalState: types.StringValue(string(node.OperationalState)),
		}
		nodeNetworks := make([]loadbalancerNodeNetworkModel, len(node.Networks))

		for j, nodeNetwork := range node.Networks {
			dataNetwork := loadbalancerNodeNetworkModel{
				Name: types.StringValue(nodeNetwork.Name),
				Type: types.StringValue(string(nodeNetwork.Type)),
			}

			nodeNetworkIPAddresses := make([]loadBalancerNodeNetworkIPAddressModel, len(nodeNetwork.IPAddresses))
			for k, nodeNetworkIPAddress := range nodeNetwork.IPAddresses {
				nodeNetworkIPAddresses[k] = loadBalancerNodeNetworkIPAddressModel{
					Address: types.StringValue(nodeNetworkIPAddress.Address),
					Listen:  types.BoolValue(nodeNetworkIPAddress.Listen),
				}
			}

			dataNodeNetworkIPAddresses, ipDiags := types.ListValueFrom(ctx, loadBalancerNodeNetworkIPAddressModel{}.ElementType(), nodeNetworkIPAddresses)
			respDiagnostics.Append(ipDiags...)
			dataNetwork.IPAddresses = dataNodeNetworkIPAddresses

			nodeNetworks[j] = dataNetwork
		}
		dataNodeNetworks, networkDiags := types.ListValueFrom(ctx, loadbalancerNodeNetworkModel{}.ElementType(), nodeNetworks)
		respDiagnostics.Append(networkDiags...)

		dataNode.Networks = dataNodeNetworks

		nodes[i] = dataNode
	}

	data.Nodes, diags = types.ListValueFrom(ctx, data.Nodes.ElementType(ctx), nodes)
	respDiagnostics.Append(diags...)

	data.OperationalState = types.StringValue(string(loadbalancer.OperationalState))
	data.Plan = types.StringValue(loadbalancer.Plan)

	resolverNames := []string{}
	for _, resolver := range loadbalancer.Resolvers {
		resolverNames = append(resolverNames, resolver.Name)
	}
	data.Resolvers, diags = types.ListValueFrom(ctx, types.StringType, resolverNames)
	respDiagnostics.Append(diags...)

	data.Zone = types.StringValue(loadbalancer.Zone)

	return respDiagnostics
}

func (r *loadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data loadBalancerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	loadBalancer, err := r.client.GetLoadBalancer(ctx, &request.GetLoadBalancerRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read loadbalancer details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setLoadBalancerValues(ctx, &data, loadBalancer)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func shouldAttachIPAddress(planIPAddress loadbalancerIPAddressModel, lbIPAddresses []upcloud.LoadBalancerFloatingIPAddress) bool {
	for _, ip := range lbIPAddresses {
		if planIPAddress.Address.ValueString() == ip.Address && planIPAddress.NetworkName.ValueString() == ip.NetworkName {
			return false
		}
	}
	return true
}

func shouldRemoveIPAddress(lbIPAddress upcloud.LoadBalancerFloatingIPAddress, dataIPAddresses []loadbalancerIPAddressModel) bool {
	for _, ip := range dataIPAddresses {
		if lbIPAddress.Address == ip.Address.ValueString() && lbIPAddress.NetworkName == ip.NetworkName.ValueString() {
			return false
		}
	}
	return true
}

func (r *loadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state loadBalancerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var dataIPAddresses []loadbalancerIPAddressModel
	resp.Diagnostics.Append(data.IPAddresses.ElementsAs(ctx, &dataIPAddresses, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lbIPAddresses, err := r.client.GetLoadBalancerIPAddresses(ctx, &request.GetLoadBalancerIPAddressesRequest{
		ServiceUUID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get load balancer IP addresses",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	// If IP addresses are configured in plan, remove IP addresses that are no longer in the plan
	if !data.IPAddresses.IsNull() {
		for _, ip := range lbIPAddresses {
			if shouldRemoveIPAddress(ip, dataIPAddresses) {
				if err := r.client.RemoveLoadBalancerIPAddress(ctx, &request.RemoveLoadBalancerIPAddressRequest{
					ServiceUUID: data.ID.ValueString(),
					Address:     ip.Address,
				}); err != nil {
					resp.Diagnostics.AddError(
						"Unable to remove IP address from the loadbalancer",
						utils.ErrorDiagnosticDetail(err),
					)
					return
				}
			}
		}

		// Wait until load balancer is in running state before continuing
		_, diags := waitForRunningState(ctx, r.client, data, "removing IP addresses")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Ensure network renaming is handled before modifying the load balancer
	if !data.Networks.Equal(state.Networks) {
		var dataNetworks []request.LoadBalancerNetwork
		resp.Diagnostics.Append(data.Networks.ElementsAs(ctx, &dataNetworks, false)...)

		var stateNetworks []request.LoadBalancerNetwork
		resp.Diagnostics.Append(state.Networks.ElementsAs(ctx, &stateNetworks, false)...)

		if resp.Diagnostics.HasError() {
			return
		}

		for i, network := range dataNetworks {
			if network.Name != stateNetworks[i].Name {
				networkAPIReq := request.ModifyLoadBalancerNetworkRequest{
					ServiceUUID: data.ID.ValueString(),
					Name:        stateNetworks[i].Name,
					Network: request.ModifyLoadBalancerNetwork{
						Name: network.Name,
					},
				}
				if _, err := r.client.ModifyLoadBalancerNetwork(ctx, &networkAPIReq); err != nil {
					resp.Diagnostics.AddError(
						"Unable to modify loadbalancer network",
						utils.ErrorDiagnosticDetail(err),
					)

					return
				}
			}
		}
	}

	// Attach IP addresses that are not in the state
	if !data.IPAddresses.IsNull() {
		for _, ip := range dataIPAddresses {
			if shouldAttachIPAddress(ip, lbIPAddresses) {
				if _, err := r.client.AttachLoadBalancerIPAddress(ctx, &request.AttachLoadBalancerIPAddressRequest{
					ServiceUUID: data.ID.ValueString(),
					Address:     ip.Address.ValueString(),
					NetworkName: ip.NetworkName.ValueString(),
				}); err != nil {
					resp.Diagnostics.AddError(
						"Unable to attach IP address to the loadbalancer",
						utils.ErrorDiagnosticDetail(err),
					)
					return
				}
			}
		}
	}

	var labelsMap map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelsMap, false)...)
	}
	labels := utils.NilAsEmptyList(utils.LabelsMapToSlice(labelsMap))

	apiReq := request.ModifyLoadBalancerRequest{
		UUID:             data.ID.ValueString(),
		Name:             data.Name.ValueString(),
		Plan:             data.Plan.ValueString(),
		ConfiguredStatus: data.ConfiguredStatus.ValueString(),
		Labels:           &labels,
		MaintenanceDOW:   upcloud.LoadBalancerMaintenanceDOW(data.MaintenanceDOW.ValueString()),
		MaintenanceTime:  data.MaintenanceTime.ValueString(),
	}

	loadBalancer, err := r.client.ModifyLoadBalancer(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(loadBalancer.UUID)

	loadBalancer, diags := waitForRunningState(ctx, r.client, data, "update")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setLoadBalancerValues(ctx, &data, loadBalancer)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *loadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data loadBalancerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLoadBalancer(ctx, &request.DeleteLoadBalancerRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	// wait before continuing so that all components destroyed
	err := r.client.WaitForLoadBalancerDeletion(ctx, &request.WaitForLoadBalancerDeletionRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for loadbalancer to be deleted",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *loadBalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
