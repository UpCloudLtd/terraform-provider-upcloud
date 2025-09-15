package network

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	validatorutil "github.com/UpCloudLtd/terraform-provider-upcloud/internal/validator"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	_ resource.Resource                = &networkResource{}
	_ resource.ResourceWithConfigure   = &networkResource{}
	_ resource.ResourceWithImportState = &networkResource{}
)

func NewNetworkResource() resource.Resource {
	return &networkResource{}
}

type networkResource struct {
	client *service.Service
}

func (r *networkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

// Configure adds the provider configured client to the resource.
func (r *networkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type networkModel struct {
	Name      types.String `tfsdk:"name"`
	ID        types.String `tfsdk:"id"`
	Type      types.String `tfsdk:"type"`
	Zone      types.String `tfsdk:"zone"`
	Router    types.String `tfsdk:"router"`
	IPNetwork types.List   `tfsdk:"ip_network"`
	Labels    types.Map    `tfsdk:"labels"`
}

type ipNetworkModel struct {
	Address                 types.String `tfsdk:"address"`
	DHCP                    types.Bool   `tfsdk:"dhcp"`
	DHCPDefaultRoute        types.Bool   `tfsdk:"dhcp_default_route"`
	DHCPDns                 types.Set    `tfsdk:"dhcp_dns"`
	DHCPRoutes              types.Set    `tfsdk:"dhcp_routes"`
	Family                  types.String `tfsdk:"family"`
	Gateway                 types.String `tfsdk:"gateway"`
	DHCPRoutesConfiguration types.Object `tfsdk:"dhcp_routes_configuration"`
}

type dhcpRoutesConfigurationModel struct {
	EffectiveRoutesAutoPopulation types.Object `tfsdk:"effective_routes_auto_population"`
}

type effectiveRoutesAutoPopulationModel struct {
	Enabled             types.Bool `tfsdk:"enabled"`
	FilterByDestination types.Set  `tfsdk:"filter_by_destination"`
	ExcludeBySource     types.Set  `tfsdk:"exclude_by_source"`
	FilterByRouteType   types.Set  `tfsdk:"filter_by_route_type"`
}

var effectiveRoutesAutoPopulationAttrTypes = map[string]attr.Type{
	"enabled":               types.BoolType,
	"filter_by_destination": types.SetType{ElemType: types.StringType},
	"exclude_by_source":     types.SetType{ElemType: types.StringType},
	"filter_by_route_type":  types.SetType{ElemType: types.StringType},
}

var dhcpRoutesConfigurationAttrTypes = map[string]attr.Type{
	"effective_routes_auto_population": types.ObjectType{
		AttrTypes: effectiveRoutesAutoPopulationAttrTypes,
	},
}

func (r *networkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an SDN private network that cloud servers and other resources from the same zone can be attached to.",
		Attributes: map[string]schema.Attribute{
			"labels": utils.LabelsAttribute("network"),
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the network.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the network.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The network type",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.StringAttribute{
				Description: "The zone the network is in, e.g. `de-fra1`. You can list available zones with `upctl zone list`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"router": schema.StringAttribute{
				Description: "UUID of a router to attach to this network.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"ip_network": schema.ListNestedBlock{
				Description: "IP subnet within the network. Network must have exactly one IP subnet.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: "The CIDR range of the subnet",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.IsCIDR),
							},
						},
						"dhcp": schema.BoolAttribute{
							Description: "Is DHCP enabled?",
							Required:    true,
						},
						"dhcp_default_route": schema.BoolAttribute{
							Description: "Is the gateway the DHCP default route?",
							Computed:    true,
							Optional:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"dhcp_dns": schema.SetAttribute{
							ElementType: types.StringType,
							Description: "The DNS servers given by DHCP",
							Computed:    true,
							Optional:    true,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(
									stringvalidator.Any(
										validatorutil.NewFrameworkStringValidator(validation.IsIPv4Address),
										validatorutil.NewFrameworkStringValidator(validation.IsIPv6Address),
									),
								),
							},
						},
						"dhcp_routes": schema.SetAttribute{
							ElementType: types.StringType,
							Description: "The additional DHCP classless static routes given by DHCP",
							Computed:    true,
							Optional:    true,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(
									dhcpRouteValidator{},
								),
							},
						},
						"family": schema.StringAttribute{
							Description: "IP address family",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(upcloud.IPAddressFamilyIPv4, upcloud.IPAddressFamilyIPv6),
							},
						},
						"gateway": schema.StringAttribute{
							Description: "Gateway address given by DHCP",
							Computed:    true,
							Optional:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},

						"dhcp_routes_configuration": schema.SingleNestedAttribute{
							Optional:    true,
							Description: "DHCP routes auto-population configuration.",
							Attributes: map[string]schema.Attribute{
								"effective_routes_auto_population": schema.SingleNestedAttribute{
									Optional:    true,
									Description: "Automatically populate effective routes.",
									Attributes: map[string]schema.Attribute{
										"enabled": schema.BoolAttribute{
											Optional:    true,
											Description: "Enable or disable route auto-population.",
										},
										"filter_by_destination": schema.SetAttribute{
											Optional:    true,
											ElementType: types.StringType,
											Description: "CIDR destinations to include when auto-populating routes.",
											Validators: []validator.Set{
												setvalidator.ValueStringsAre(
													validatorutil.NewFrameworkStringValidator(validation.IsCIDR),
												),
											},
										},
										"exclude_by_source": schema.SetAttribute{
											Optional:    true,
											ElementType: types.StringType,
											Description: "Exclude routes coming from specific sources (router-connected-networks, static-route).",
											Validators: []validator.Set{
												setvalidator.ValueStringsAre(
													stringvalidator.OneOf("router-connected-networks", "static-route"),
												),
											},
										},
										"filter_by_route_type": schema.SetAttribute{
											Optional:    true,
											ElementType: types.StringType,
											Description: "Include only routes of given types (service, user).",
											Validators: []validator.Set{
												setvalidator.ValueStringsAre(
													stringvalidator.OneOf("service", "user"),
												),
											},
										},
									},
								},
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, 1),
				},
			},
		},
	}
}

// detectDHCPInputShape inspects the user's input (plan during Create/Update; state during Read)
// and returns:
//   - outerPresent: whether dhcp_routes_configuration was provided at all
//   - innerProvided: whether effective_routes_auto_population key was present
//   - innerExplicitEmpty: whether inner was present but empty (i.e., enabled unset => null/unknown)
func detectDHCPInputShape(ctx context.Context, obj types.Object) (outerPresent bool, innerProvided bool, innerExplicitEmpty bool, diags diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return false, false, false, diags
	}
	var cfg dhcpRoutesConfigurationModel
	d := obj.As(ctx, &cfg, basetypes.ObjectAsOptions{})
	diags.Append(d...)

	if cfg.EffectiveRoutesAutoPopulation.IsNull() || cfg.EffectiveRoutesAutoPopulation.IsUnknown() {
		// outer present, inner NOT provided at all (outer = {})
		return true, false, false, diags
	}

	var era effectiveRoutesAutoPopulationModel
	d2 := cfg.EffectiveRoutesAutoPopulation.As(ctx, &era, basetypes.ObjectAsOptions{})
	diags.Append(d2...)

	hasEnabled := !(era.Enabled.IsNull() || era.Enabled.IsUnknown())
	hasFBD := !(era.FilterByDestination.IsNull() || era.FilterByDestination.IsUnknown())
	hasEBS := !(era.ExcludeBySource.IsNull() || era.ExcludeBySource.IsUnknown())
	hasFBRT := !(era.FilterByRouteType.IsNull() || era.FilterByRouteType.IsUnknown())

	// inner is explicitly empty only if none of the inner attributes were provided
	return true, true, !(hasEnabled || hasFBD || hasEBS || hasFBRT), diags
}

func setValues(ctx context.Context, data *networkModel, network *upcloud.Network) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	data.Name = types.StringValue(network.Name)
	data.ID = types.StringValue(network.UUID)
	data.Type = types.StringValue(network.Type)
	data.Zone = types.StringValue(network.Zone)

	data.Labels, respDiagnostics = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(network.Labels))

	if network.Router == "" {
		data.Router = types.StringNull()
	} else {
		data.Router = types.StringValue(network.Router)
	}

	ipNetworks := make([]ipNetworkModel, len(network.IPNetworks))

	// Read the current input (Plan during Create/Update; State during Read)
	var inputIPNets []ipNetworkModel
	if !data.IPNetwork.IsNull() && !data.IPNetwork.IsUnknown() {
		diags := data.IPNetwork.ElementsAs(ctx, &inputIPNets, false)
		respDiagnostics.Append(diags...)
	}

	for i, ipnet := range network.IPNetworks {
		ipNetworks[i].Address = types.StringValue(ipnet.Address)
		ipNetworks[i].DHCP = utils.AsBool(ipnet.DHCP)
		ipNetworks[i].DHCPDefaultRoute = utils.AsBool(ipnet.DHCPDefaultRoute)

		dhcpdns, diags := types.SetValueFrom(ctx, types.StringType, utils.NilAsEmptyList(ipnet.DHCPDns))
		respDiagnostics.Append(diags...)
		ipNetworks[i].DHCPDns = dhcpdns

		dhcproutes, diags := types.SetValueFrom(ctx, types.StringType, utils.NilAsEmptyList(ipnet.DHCPRoutes))
		respDiagnostics.Append(diags...)
		ipNetworks[i].DHCPRoutes = dhcproutes

		ipNetworks[i].Family = types.StringValue(ipnet.Family)
		ipNetworks[i].Gateway = types.StringValue(ipnet.Gateway)

		// Figure out the user's intended shape for dhcp_routes_configuration
		var outerPresent, innerProvided, innerExplicitEmpty bool
		if len(inputIPNets) > i {
			op, ip, iee, d := detectDHCPInputShape(ctx, inputIPNets[i].DHCPRoutesConfiguration)
			respDiagnostics.Append(d...)
			outerPresent, innerProvided, innerExplicitEmpty = op, ip, iee
		}

		// API always returns ERA with enabled defaulting to false.
		era := ipnet.DHCPRoutesConfiguration.EffectiveRoutesAutoPopulation
		enabledTF := utils.AsBool(era.Enabled)

		var fbdStrings []string
		if era.FilterByDestination != nil {
			fbdStrings = *era.FilterByDestination
		}
		fbdSetFromAPI, diags := types.SetValueFrom(ctx, types.StringType, utils.NilAsEmptyList(fbdStrings))
		respDiagnostics.Append(diags...)

		var ebsStrings []string
		if era.ExcludeBySource != nil {
			ebsStrings = make([]string, len(*era.ExcludeBySource))
			for i, v := range *era.ExcludeBySource {
				ebsStrings[i] = string(v)
			}
		}
		ebsSetFromAPI, diags := types.SetValueFrom(ctx, types.StringType, utils.NilAsEmptyList(ebsStrings))

		var frtStrings []string
		if era.FilterByRouteType != nil {
			frtStrings = make([]string, len(*era.FilterByRouteType))
			for i, v := range *era.FilterByRouteType {
				frtStrings[i] = string(v)
			}
		}
		frtSetFromAPI, diags := types.SetValueFrom(ctx, types.StringType, utils.NilAsEmptyList(frtStrings))

		switch {
		case !outerPresent:
			ipNetworks[i].DHCPRoutesConfiguration = types.ObjectNull(dhcpRoutesConfigurationAttrTypes)

		case outerPresent && !innerProvided:
			cfgObj, d := types.ObjectValue(
				dhcpRoutesConfigurationAttrTypes,
				map[string]attr.Value{
					"effective_routes_auto_population": types.ObjectNull(effectiveRoutesAutoPopulationAttrTypes),
				},
			)
			respDiagnostics.Append(d...)
			ipNetworks[i].DHCPRoutesConfiguration = cfgObj

		case outerPresent && innerProvided && innerExplicitEmpty:
			// inner {}  -> enabled=null, filter_by_destination=null etc
			eraObj, d1 := types.ObjectValue(
				effectiveRoutesAutoPopulationAttrTypes,
				map[string]attr.Value{
					"enabled":               types.BoolNull(),
					"filter_by_destination": types.SetNull(types.StringType),
					"exclude_by_source":     types.SetNull(types.StringType),
					"filter_by_route_type":  types.SetNull(types.StringType),
				},
			)
			respDiagnostics.Append(d1...)
			cfgObj, d2 := types.ObjectValue(
				dhcpRoutesConfigurationAttrTypes,
				map[string]attr.Value{
					"effective_routes_auto_population": eraObj,
				},
			)
			respDiagnostics.Append(d2...)
			ipNetworks[i].DHCPRoutesConfiguration = cfgObj

		default:
			// inner provided with some attrs; preserve per-attribute shape from input
			var inEra effectiveRoutesAutoPopulationModel
			// We know innerProvided == true; decode the input inner to see which attrs were present
			if len(inputIPNets) > i {
				var inCfg dhcpRoutesConfigurationModel
				d := inputIPNets[i].DHCPRoutesConfiguration.As(ctx, &inCfg, basetypes.ObjectAsOptions{})
				respDiagnostics.Append(d...)
				d = inCfg.EffectiveRoutesAutoPopulation.As(ctx, &inEra, basetypes.ObjectAsOptions{})
				respDiagnostics.Append(d...)
			}

			// enabled: null in state if absent in input; else API value
			var enabledAttr attr.Value
			if inEra.Enabled.IsNull() || inEra.Enabled.IsUnknown() {
				enabledAttr = types.BoolNull()
			} else {
				enabledAttr = enabledTF
			}

			// exclude_by_source
			var ebsAttr attr.Value
			if inEra.ExcludeBySource.IsNull() || inEra.ExcludeBySource.IsUnknown() {
				ebsAttr = types.SetNull(types.StringType)
			} else {
				ebsAttr = ebsSetFromAPI
			}

			// filter_by_route_type
			var frtAttr attr.Value
			if inEra.FilterByRouteType.IsNull() || inEra.FilterByRouteType.IsUnknown() {
				frtAttr = types.SetNull(types.StringType)
			} else {
				frtAttr = frtSetFromAPI
			}

			// filter_by_destination: null in state if absent in input; else API value (empty set if provided as [])
			var fbdAttr attr.Value
			if inEra.FilterByDestination.IsNull() || inEra.FilterByDestination.IsUnknown() {
				fbdAttr = types.SetNull(types.StringType)
			} else {
				fbdAttr = fbdSetFromAPI
			}

			eraObj, d1 := types.ObjectValue(
				effectiveRoutesAutoPopulationAttrTypes,
				map[string]attr.Value{
					"enabled":               enabledAttr,
					"filter_by_destination": fbdAttr,
					"exclude_by_source":     ebsAttr,
					"filter_by_route_type":  frtAttr,
				},
			)
			respDiagnostics.Append(d1...)

			cfgObj, d2 := types.ObjectValue(
				dhcpRoutesConfigurationAttrTypes,
				map[string]attr.Value{
					"effective_routes_auto_population": eraObj,
				},
			)
			respDiagnostics.Append(d2...)
			ipNetworks[i].DHCPRoutesConfiguration = cfgObj
		}
	}

	var diags diag.Diagnostics
	data.IPNetwork, diags = types.ListValueFrom(ctx, data.IPNetwork.ElementType(ctx), ipNetworks)
	respDiagnostics.Append(diags...)
	return respDiagnostics
}

func buildIPNetworks(ctx context.Context, dataIPNetworks types.List) ([]upcloud.IPNetwork, diag.Diagnostics) {
	var planNetworks []ipNetworkModel
	respDiagnostics := dataIPNetworks.ElementsAs(ctx, &planNetworks, false)

	networks := make([]upcloud.IPNetwork, 0, len(planNetworks))

	for _, ipnet := range planNetworks {
		dhcpdns, diags := utils.SetAsSliceOfStrings(ctx, ipnet.DHCPDns)
		respDiagnostics.Append(diags...)

		dhcproutes, diags := utils.SetAsSliceOfStrings(ctx, ipnet.DHCPRoutes)
		respDiagnostics.Append(diags...)

		ipNet := upcloud.IPNetwork{
			Address:          ipnet.Address.ValueString(),
			DHCP:             utils.AsUpCloudBoolean(ipnet.DHCP),
			DHCPDefaultRoute: utils.AsUpCloudBoolean(ipnet.DHCPDefaultRoute),
			DHCPDns:          dhcpdns,
			DHCPRoutes:       dhcproutes,
			Family:           ipnet.Family.ValueString(),
			Gateway:          ipnet.Gateway.ValueString(),
		}

		if ipnet.DHCPRoutesConfiguration.IsNull() || ipnet.DHCPRoutesConfiguration.IsUnknown() {
			// Outer removed / not set: explicitly clear server-side config.
			ipNet.DHCPRoutesConfiguration = upcloud.DHCPRoutesConfiguration{
				EffectiveRoutesAutoPopulation: upcloud.EffectiveRoutesAutoPopulation{
					Enabled:             upcloud.FromBool(false),
					FilterByDestination: &[]string{},
					ExcludeBySource:     &[]upcloud.NetworkRouteSource{},
					FilterByRouteType:   &[]upcloud.NetworkRouteType{},
				},
			}
		} else {
			var cfg dhcpRoutesConfigurationModel
			diags := ipnet.DHCPRoutesConfiguration.As(ctx, &cfg, basetypes.ObjectAsOptions{})
			respDiagnostics.Append(diags...)

			if cfg.EffectiveRoutesAutoPopulation.IsNull() || cfg.EffectiveRoutesAutoPopulation.IsUnknown() {
				// Inner missing: clear.
				ipNet.DHCPRoutesConfiguration = upcloud.DHCPRoutesConfiguration{
					EffectiveRoutesAutoPopulation: upcloud.EffectiveRoutesAutoPopulation{
						Enabled:             upcloud.FromBool(false),
						FilterByDestination: &[]string{},
						ExcludeBySource:     &[]upcloud.NetworkRouteSource{},
						FilterByRouteType:   &[]upcloud.NetworkRouteType{},
					},
				}
			} else {
				var era effectiveRoutesAutoPopulationModel
				diags := cfg.EffectiveRoutesAutoPopulation.As(ctx, &era, basetypes.ObjectAsOptions{})
				respDiagnostics.Append(diags...)

				// enabled: unknown/null => false
				var enabledUC upcloud.Boolean
				if era.Enabled.IsUnknown() || era.Enabled.IsNull() {
					enabledUC = upcloud.FromBool(false)
				} else {
					enabledUC = utils.AsUpCloudBoolean(era.Enabled)
				}

				// filter_by_destination: null/unknown => empty slice (clear)
				fbd, d3 := utils.SetAsSliceOfStrings(ctx, era.FilterByDestination)
				respDiagnostics.Append(d3...)
				fbd = utils.NilAsEmptyList(fbd)

				// exclude_by_source
				ebsStrings, d4 := utils.SetAsSliceOfStrings(ctx, era.ExcludeBySource)
				respDiagnostics.Append(d4...)
				ebsStrings = utils.NilAsEmptyList(ebsStrings)
				ebs := make([]upcloud.NetworkRouteSource, 0, len(ebsStrings))
				for _, s := range ebsStrings {
					ebs = append(ebs, upcloud.NetworkRouteSource(s))
				}

				// filter_by_route_type
				frtStrings, d5 := utils.SetAsSliceOfStrings(ctx, era.FilterByRouteType)
				respDiagnostics.Append(d5...)
				frtStrings = utils.NilAsEmptyList(frtStrings)
				frt := make([]upcloud.NetworkRouteType, 0, len(frtStrings))
				for _, s := range frtStrings {
					frt = append(frt, upcloud.NetworkRouteType(s))
				}

				ipNet.DHCPRoutesConfiguration = upcloud.DHCPRoutesConfiguration{
					EffectiveRoutesAutoPopulation: upcloud.EffectiveRoutesAutoPopulation{
						Enabled:             enabledUC,
						FilterByDestination: &fbd,
						ExcludeBySource:     &ebs,
						FilterByRouteType:   &frt,
					},
				}
			}
		}

		networks = append(networks, ipNet)
	}

	return networks, respDiagnostics
}

func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data networkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}

	apiReq := request.CreateNetworkRequest{
		Name:   data.Name.ValueString(),
		Labels: utils.LabelsMapToSlice(labels),
		Zone:   data.Zone.ValueString(),
		Router: data.Router.ValueString(),
	}

	networks, diags := buildIPNetworks(ctx, data.IPNetwork)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiReq.IPNetworks = networks

	network, err := r.client.CreateNetwork(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create network",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, network)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data networkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	network, err := r.client.GetNetworkDetails(ctx, &request.GetNetworkDetailsRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read network details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, network)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data networkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}
	labelsSlice := utils.NilAsEmptyList(utils.LabelsMapToSlice(labels))

	apiReq := request.ModifyNetworkRequest{
		UUID:   data.ID.ValueString(),
		Name:   data.Name.ValueString(),
		Labels: &labelsSlice,
	}

	networks, diags := buildIPNetworks(ctx, data.IPNetwork)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiReq.IPNetworks = networks

	network, err := r.client.ModifyNetwork(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify network",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if network.Router != data.Router.ValueString() {
		err = r.client.AttachNetworkRouter(ctx, &request.AttachNetworkRouterRequest{
			NetworkUUID: data.ID.ValueString(),
			RouterUUID:  data.Router.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to modify networks router attachment",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}

		network, err = r.client.GetNetworkDetails(ctx, &request.GetNetworkDetailsRequest{
			UUID: data.ID.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read network details",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
	}

	resp.Diagnostics.Append(setValues(ctx, &data, network)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data networkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if err := r.client.DeleteNetwork(ctx, &request.DeleteNetworkRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete network",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *networkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
