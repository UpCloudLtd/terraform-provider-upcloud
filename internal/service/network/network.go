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
}

type ipNetworkModel struct {
	Address          types.String `tfsdk:"address"`
	DHCP             types.Bool   `tfsdk:"dhcp"`
	DHCPDefaultRoute types.Bool   `tfsdk:"dhcp_default_route"`
	DHCPDns          types.Set    `tfsdk:"dhcp_dns"`
	DHCPRoutes       types.Set    `tfsdk:"dhcp_routes"`
	Family           types.String `tfsdk:"family"`
	Gateway          types.String `tfsdk:"gateway"`
}

func (r *networkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an SDN private network that cloud servers and other resources from the same zone can be attached to.",
		Attributes: map[string]schema.Attribute{
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
									stringvalidator.Any(
										validatorutil.NewFrameworkStringValidator(validation.IsCIDR),
									),
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

func setValues(ctx context.Context, data *networkModel, network *upcloud.Network) diag.Diagnostics {
	respDiagnostics := diag.Diagnostics{}

	data.Name = types.StringValue(network.Name)
	data.ID = types.StringValue(network.UUID)
	data.Type = types.StringValue(network.Type)
	data.Zone = types.StringValue(network.Zone)

	if network.Router == "" {
		data.Router = types.StringNull()
	} else {
		data.Router = types.StringValue(network.Router)
	}

	ipNetworks := make([]ipNetworkModel, len(network.IPNetworks))

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
	}

	var diags diag.Diagnostics
	data.IPNetwork, diags = types.ListValueFrom(ctx, data.IPNetwork.ElementType(ctx), ipNetworks)
	respDiagnostics.Append(diags...)

	return respDiagnostics
}

func buildIPNetworks(ctx context.Context, dataIPNetworks types.List) ([]upcloud.IPNetwork, diag.Diagnostics) {
	var planNetworks []ipNetworkModel
	respDiagnostics := dataIPNetworks.ElementsAs(ctx, &planNetworks, false)

	networks := make([]upcloud.IPNetwork, 0)

	for _, ipnet := range planNetworks {
		dhcpdns, diags := utils.SetAsSliceOfStrings(ctx, ipnet.DHCPDns)
		respDiagnostics.Append(diags...)

		dhcproutes, diags := utils.SetAsSliceOfStrings(ctx, ipnet.DHCPRoutes)
		respDiagnostics.Append(diags...)

		networks = append(networks, upcloud.IPNetwork{
			Address:          ipnet.Address.ValueString(),
			DHCP:             utils.AsUpCloudBoolean(ipnet.DHCP),
			DHCPDefaultRoute: utils.AsUpCloudBoolean(ipnet.DHCPDefaultRoute),
			DHCPDns:          dhcpdns,
			DHCPRoutes:       dhcproutes,
			Family:           ipnet.Family.ValueString(),
			Gateway:          ipnet.Gateway.ValueString(),
		})
	}

	return networks, respDiagnostics
}

func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data networkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := request.CreateNetworkRequest{
		Name:   data.Name.ValueString(),
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

	network, err := r.client.GetNetworkDetails(ctx, &request.GetNetworkDetailsRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read network details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, network)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data networkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	apiReq := request.ModifyNetworkRequest{
		UUID: data.ID.ValueString(),
		Name: data.Name.ValueString(),
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
			"Unable to delete network",
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
