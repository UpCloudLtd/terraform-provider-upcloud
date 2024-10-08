package router

import (
	"context"
	"sort"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	validatorutil "github.com/UpCloudLtd/terraform-provider-upcloud/internal/validator"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	_ resource.Resource                = &routerResource{}
	_ resource.ResourceWithConfigure   = &routerResource{}
	_ resource.ResourceWithImportState = &routerResource{}
)

func NewRouterResource() resource.Resource {
	return &routerResource{}
}

type routerResource struct {
	client *service.Service
}

func (r *routerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_router"
}

// Configure adds the provider configured client to the resource.
func (r *routerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type routerModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	AttachedNetworks types.List   `tfsdk:"attached_networks"`
	UserStaticRoutes types.Set    `tfsdk:"static_route"`
	StaticRoutes     types.Set    `tfsdk:"static_routes"`
	Labels           types.Map    `tfsdk:"labels"`
}

type staticRouteModel struct {
	Name    types.String `tfsdk:"name"`
	Nexthop types.String `tfsdk:"nexthop"`
	Route   types.String `tfsdk:"route"`
	Type    types.String `tfsdk:"type"`
}

var staticRouteTypes = map[string]attr.Type{
	"name":    types.StringType,
	"nexthop": types.StringType,
	"route":   types.StringType,
	"type":    types.StringType,
}

func (r *routerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Routers can be used to connect multiple Private Networks. UpCloud Servers on any attached network can communicate directly with each other.",
		Attributes: map[string]schema.Attribute{
			"labels": utils.LabelsAttribute("router"),
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the router.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the router.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the router",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"attached_networks": schema.ListAttribute{
				Description: "List of UUIDs representing networks attached to this router.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"static_routes": schema.SetAttribute{
				MarkdownDescription: "A collection of static routes for this router. This set includes both user and service defined static routes. The objects in this set use the same schema as `static_route` blocks.",
				Computed:            true,
				ElementType:         types.ObjectType{AttrTypes: staticRouteTypes},
			},
		},
		Blocks: map[string]schema.Block{
			"static_route": schema.SetNestedBlock{
				Description: "A collection of user managed static routes for this router.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name or description of the route.",
							Optional:    true,
							Computed:    true,
						},
						"nexthop": schema.StringAttribute{
							Description: "Next hop address. NOTE: For static route to be active the next hop has to be an address of a reachable running Cloud Server in one of the Private Networks attached to the router.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.Any(
									validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsIPv4Address, validation.IsIPv6Address)),
									stringvalidator.OneOf("no-nexthop"),
								),
							},
						},
						"route": schema.StringAttribute{
							Description: "Destination prefix of the route.",
							Required:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.IsCIDR),
							},
						},
						"type": schema.StringAttribute{
							Description: "Type of the route.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func setValues(ctx context.Context, data *routerModel, router *upcloud.Router) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	data.Name = types.StringValue(router.Name)
	data.ID = types.StringValue(router.UUID)
	data.Type = types.StringValue(router.Type)

	attachedNetworkUUIDs := make([]string, 0)
	for _, network := range router.AttachedNetworks {
		attachedNetworkUUIDs = append(attachedNetworkUUIDs, network.NetworkUUID)
	}
	sort.Strings(attachedNetworkUUIDs)

	attachedNetworks, diags := types.ListValueFrom(ctx, types.StringType, utils.NilAsEmptyList(attachedNetworkUUIDs))
	respDiagnostics.Append(diags...)
	data.AttachedNetworks = attachedNetworks

	data.Labels, respDiagnostics = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(router.Labels))

	staticRoutes := make([]staticRouteModel, 0)
	userStaticRoutes := make([]staticRouteModel, 0)
	for _, route := range router.StaticRoutes {
		r := staticRouteModel{
			Name:    types.StringValue(route.Name),
			Nexthop: types.StringValue(route.Nexthop),
			Route:   types.StringValue(route.Route),
			Type:    types.StringValue(string(route.Type)),
		}
		staticRoutes = append(staticRoutes, r)

		if route.Type == upcloud.RouterStaticRouteTypeUser {
			userStaticRoutes = append(userStaticRoutes, r)
		}
	}

	data.StaticRoutes, diags = types.SetValueFrom(ctx, data.StaticRoutes.ElementType(ctx), staticRoutes)
	respDiagnostics.Append(diags...)

	data.UserStaticRoutes, diags = types.SetValueFrom(ctx, data.UserStaticRoutes.ElementType(ctx), userStaticRoutes)
	respDiagnostics.Append(diags...)

	return respDiagnostics
}

func buildStaticRoutes(ctx context.Context, dataUserStaticRoutes types.Set) ([]upcloud.StaticRoute, diag.Diagnostics) {
	var planUserStaticRoutes []staticRouteModel
	respDiagnostics := dataUserStaticRoutes.ElementsAs(ctx, &planUserStaticRoutes, false)

	staticRoutes := make([]upcloud.StaticRoute, 0)

	for _, route := range planUserStaticRoutes {
		staticRoutes = append(staticRoutes, upcloud.StaticRoute{
			Name:    route.Name.ValueString(),
			Nexthop: route.Nexthop.ValueString(),
			Route:   route.Route.ValueString(),
		})
	}

	return staticRoutes, respDiagnostics
}

func (r *routerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data routerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}

	staticRoutes, diags := buildStaticRoutes(ctx, data.UserStaticRoutes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := request.CreateRouterRequest{
		Name:         data.Name.ValueString(),
		Labels:       utils.LabelsMapToSlice(labels),
		StaticRoutes: staticRoutes,
	}

	router, err := r.client.CreateRouter(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create router",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, router)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *routerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data routerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	router, err := r.client.GetRouterDetails(ctx, &request.GetRouterDetailsRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read router details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, router)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *routerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data routerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}
	labelsSlice := utils.NilAsEmptyList(utils.LabelsMapToSlice(labels))

	staticRoutes, diags := buildStaticRoutes(ctx, data.UserStaticRoutes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := request.ModifyRouterRequest{
		UUID:         data.ID.ValueString(),
		Name:         data.Name.ValueString(),
		Labels:       &labelsSlice,
		StaticRoutes: &staticRoutes,
	}

	router, err := r.client.ModifyRouter(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify router",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, router)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *routerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data routerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	router, err := r.client.GetRouterDetails(ctx, &request.GetRouterDetailsRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read router",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	if len(router.AttachedNetworks) > 0 {
		for _, network := range router.AttachedNetworks {
			err := r.client.DetachNetworkRouter(ctx, &request.DetachNetworkRouterRequest{
				NetworkUUID: network.NetworkUUID,
			})
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to detach router from a network",
					utils.ErrorDiagnosticDetail(err),
				)
			}
		}
	}

	if err := r.client.DeleteRouter(ctx, &request.DeleteRouterRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete router",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *routerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
