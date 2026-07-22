package gateway

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	idDescription               = "Gateway UUID."
	nameDescription             = "Gateway name. Needs to be unique within the account."
	zoneDescription             = "Zone in which the gateway will be hosted, e.g. `de-fra1`."
	featuresDescription         = "Features enabled for the gateway. Valid item values are `nat` and `vpn`. For more details, see documentation on [NAT](https://upcloud.com/docs/products/nat-gateway/) and [VPN](https://upcloud.com/docs/products/vpn-gateway/) gateways."
	routerDescription           = "Attached Router from where traffic is routed towards the network gateway service."
	routerIDDescription         = "ID of the router attached to the gateway."
	configuredStatusDescription = "The service configured status indicates the service's current intended status. Managed by the customer."
	operationalStateDescription = "The service operational state indicates the service's current operational, effective state. Managed by the system."
	addressesDescription        = "IP addresses assigned to the gateway."
	planDescription             = "Gateway pricing plan, defaults to `development`. You can list available plans with `upctl gateway plans`."
	connectionsDescription      = "Names of connections attached to the gateway. Note that this field can have outdated information as connections are created by a separate resource. To make sure that you have the most recent data run 'terraform refresh'."

	cleanupWaitTimeSeconds = 15
	resourceNameRegexpStr  = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
)

var resourceNameRegexp = regexp.MustCompile(resourceNameRegexpStr)

var (
	_ resource.Resource                = &gatewayResource{}
	_ resource.ResourceWithConfigure   = &gatewayResource{}
	_ resource.ResourceWithImportState = &gatewayResource{}
)

func NewGatewayResource() resource.Resource {
	return &gatewayResource{}
}

type gatewayResource struct {
	client *service.Service
}

func (r *gatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway"
}

// Configure adds the provider configured client to the resource.
func (r *gatewayResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type addressModel struct {
	Address types.String `tfsdk:"address"`
	Name    types.String `tfsdk:"name"`
}

type routerModel struct {
	ID types.String `tfsdk:"id"`
}

type gatewayModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Zone             types.String `tfsdk:"zone"`
	Features         types.Set    `tfsdk:"features"`
	Labels           types.Map    `tfsdk:"labels"`
	ConfiguredStatus types.String `tfsdk:"configured_status"`
	OperationalState types.String `tfsdk:"operational_state"`
	Plan             types.String `tfsdk:"plan"`
	Connections      types.List   `tfsdk:"connections"`
	Addresses        types.Set    `tfsdk:"addresses"`
	Router           types.Set    `tfsdk:"router"`
	Address          types.Set    `tfsdk:"address"`
}

func (r *gatewayResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Network gateways connect SDN Private Networks to external IP networks.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription,
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(resourceNameRegexp, fmt.Sprintf("must contain only alphanumeric characters, hyphens, and underscores (a-z, 0-9, -, _). Regular expresion used to check validation: %s", resourceNameRegexp)),
				},
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: zoneDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"features": schema.SetAttribute{
				MarkdownDescription: featuresDescription,
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							string(upcloud.GatewayFeatureNAT),
							string(upcloud.GatewayFeatureVPN),
						),
					),
				},
			},
			"labels": utils.LabelsAttribute("gateway"),
			"configured_status": schema.StringAttribute{
				MarkdownDescription: configuredStatusDescription,
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(string(upcloud.GatewayConfiguredStatusStarted)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(upcloud.GatewayConfiguredStatusStarted),
						string(upcloud.GatewayConfiguredStatusStopped),
					),
				},
			},
			"operational_state": schema.StringAttribute{
				MarkdownDescription: operationalStateDescription,
				Computed:            true,
			},
			"plan": schema.StringAttribute{
				MarkdownDescription: planDescription,
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString("development"),
			},
			"connections": schema.ListAttribute{
				MarkdownDescription: connectionsDescription,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"addresses": schema.SetNestedAttribute{
				MarkdownDescription: "Use 'address' attribute instead. This attribute will be removed in the next major version of the provider.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: "IP address.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the address.",
							Computed:    true,
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"router": schema.SetNestedBlock{
				MarkdownDescription: routerDescription,
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: routerIDDescription,
							Required:            true,
						},
					},
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.SizeAtMost(1),
				},
			},
			"address": schema.SetNestedBlock{
				MarkdownDescription: addressesDescription,
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: "IP address.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the IP address.",
							Computed:    true,
							Optional:    true,
						},
					},
				},
				Validators: []validator.Set{
					setvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func setGatewayValues(ctx context.Context, data *gatewayModel, gw *upcloud.Gateway) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	isImport := data.Name.ValueString() == ""

	data.ID = types.StringValue(gw.UUID)
	data.Name = types.StringValue(gw.Name)
	data.Zone = types.StringValue(gw.Zone)
	data.Plan = types.StringValue(gw.Plan)
	data.ConfiguredStatus = types.StringValue(string(gw.ConfiguredStatus))
	data.OperationalState = types.StringValue(string(gw.OperationalState))

	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(gw.Labels))
	respDiagnostics.Append(diags...)

	features := make([]string, len(gw.Features))
	for i, feature := range gw.Features {
		features[i] = string(feature)
	}

	data.Features, diags = types.SetValueFrom(ctx, data.Features.ElementType(ctx), features)
	respDiagnostics.Append(diags...)

	connections := make([]string, len(gw.Connections))
	for i, connection := range gw.Connections {
		connections[i] = connection.Name
	}

	data.Connections, diags = types.ListValueFrom(ctx, data.Connections.ElementType(ctx), connections)
	respDiagnostics.Append(diags...)

	routers := make([]routerModel, len(gw.Routers))
	for i, router := range gw.Routers {
		routers[i].ID = types.StringValue(router.UUID)
	}

	data.Router, diags = types.SetValueFrom(ctx, data.Router.ElementType(ctx), routers)
	respDiagnostics.Append(diags...)

	addresses := make([]addressModel, len(gw.Addresses))
	for i, address := range gw.Addresses {
		addresses[i].Address = types.StringValue(address.Address)
		addresses[i].Name = types.StringValue(address.Name)
	}

	data.Addresses, diags = types.SetValueFrom(ctx, data.Addresses.ElementType(ctx), addresses)
	respDiagnostics.Append(diags...)

	if !data.Address.IsNull() || isImport {
		data.Address, diags = types.SetValueFrom(ctx, data.Address.ElementType(ctx), addresses)
		respDiagnostics.Append(diags...)
	}

	return respDiagnostics
}

func (r *gatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data gatewayModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}

	var features []upcloud.GatewayFeature
	resp.Diagnostics.Append(data.Features.ElementsAs(ctx, &features, false)...)

	var routers []routerModel
	resp.Diagnostics.Append(data.Router.ElementsAs(ctx, &routers, false)...)

	apiReq := request.CreateGatewayRequest{
		Name:     data.Name.ValueString(),
		Zone:     data.Zone.ValueString(),
		Plan:     data.Plan.ValueString(),
		Features: features,
		Routers: []request.GatewayRouter{{
			UUID: routers[0].ID.ValueString(),
		}},
		Labels:           utils.LabelsMapToSlice(labels),
		ConfiguredStatus: upcloud.GatewayConfiguredStatus(data.ConfiguredStatus.ValueString()),
	}

	addresses := []upcloud.GatewayAddress{}
	if !data.Address.IsNull() && !data.Address.IsUnknown() {
		var addressesData []addressModel
		resp.Diagnostics.Append(data.Address.ElementsAs(ctx, &addressesData, false)...)

		for _, address := range addressesData {
			if address.Address.ValueString() == "" {
				resp.Diagnostics.AddError(
					"Unable to build create gateway request",
					"Gateway address name cannot be empty.",
				)
				return
			}

			addresses = append(addresses, upcloud.GatewayAddress{
				Name: address.Name.ValueString(),
			})
		}
	}

	if len(addresses) > 0 {
		apiReq.Addresses = addresses
	}

	gw, err := r.client.CreateGateway(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create gateway",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	data.ID = types.StringValue(gw.UUID)

	gw, err = waitForGatewayToBeRunning(ctx, r.client, gw.UUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while waiting for gateway to be in running state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setGatewayValues(ctx, &data, gw)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data gatewayModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	gw, err := r.client.GetGateway(ctx, &request.GetGatewayRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read gateway details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setGatewayValues(ctx, &data, gw)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan gatewayModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	uuid := plan.ID.ValueString()

	var labels map[string]string
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	}
	labelsSlice := utils.NilAsEmptyList(utils.LabelsMapToSlice(labels))

	apiReq := &request.ModifyGatewayRequest{
		UUID:             uuid,
		Name:             plan.Name.ValueString(),
		Plan:             plan.Plan.ValueString(),
		ConfiguredStatus: upcloud.GatewayConfiguredStatus(plan.ConfiguredStatus.ValueString()),
		Labels:           labelsSlice,
	}

	_, err := r.client.ModifyGateway(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify gateway",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	gw, err := waitForGatewayToBeRunning(ctx, r.client, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while waiting for gateway to be in running state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setGatewayValues(ctx, &plan, gw)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data gatewayModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if err := r.client.DeleteGateway(ctx, &request.DeleteGatewayRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete gateway",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	err := waitForGatewayToBeDeleted(ctx, r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while waiting for gateway to be deleted",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	// Additionally wait some time so that all cleanup operations can finish
	time.Sleep(time.Second * cleanupWaitTimeSeconds)
}

func (r *gatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitForGatewayToBeRunning(ctx context.Context, svc *service.Service, id string) (*upcloud.Gateway, error) {
	const maxRetries int = 500

	for i := 0; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			gw, err := svc.GetGateway(ctx, &request.GetGatewayRequest{UUID: id})
			if err != nil {
				return nil, err
			}
			if gw.OperationalState == upcloud.GatewayOperationalStateRunning {
				return gw, nil
			}

			tflog.Info(ctx, "waiting for network gateway to be running", map[string]interface{}{"name": gw.Name, "state": gw.OperationalState})
		}
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("max retries (%d)reached while waiting for network gateway to be running", maxRetries)
}

func getGatewayDeleted(ctx context.Context, svc *service.Service, id ...string) (map[string]interface{}, error) {
	gw, err := svc.GetGateway(ctx, &request.GetGatewayRequest{UUID: id[0]})

	return map[string]interface{}{"resource": "gateway", "name": gw.Name, "state": gw.OperationalState}, err
}

func waitForGatewayToBeDeleted(ctx context.Context, svc *service.Service, id string) error {
	return utils.WaitForResourceToBeDeleted(ctx, svc, getGatewayDeleted, id)
}
