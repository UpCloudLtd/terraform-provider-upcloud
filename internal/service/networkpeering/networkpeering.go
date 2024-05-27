package networkpeering

import (
	"context"
	"fmt"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &networkPeeringResource{}
	_ resource.ResourceWithConfigure   = &networkPeeringResource{}
	_ resource.ResourceWithImportState = &networkPeeringResource{}
)

func NewNetworkPeeringResource() resource.Resource {
	return &networkPeeringResource{}
}

type networkPeeringResource struct {
	client *service.Service
}

func (r *networkPeeringResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_peering"
}

// Configure adds the provider configured client to the resource.
func (r *networkPeeringResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type networkPeeringModel struct {
	ConfiguredStatus types.String   `tfsdk:"configured_status"`
	Labels           types.Map      `tfsdk:"labels"`
	Name             types.String   `tfsdk:"name"`
	Network          []networkModel `tfsdk:"network"`
	PeerNetwork      []networkModel `tfsdk:"peer_network"`
	ID               types.String   `tfsdk:"id"`
}

type networkModel struct {
	UUID types.String `tfsdk:"uuid"`
}

func networkBlock(kind string) schema.Block {
	return &schema.ListNestedBlock{
		Description: fmt.Sprintf("%s network of the network peering.", kind),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"uuid": schema.StringAttribute{
					Description: "The UUID of the network.",
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeBetween(1, 1),
		},
	}
}

func (r *networkPeeringResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Network peerings can be used to connect networks across accounts. For the network peering to become active, the peering must be made from both directions.",
		Attributes: map[string]schema.Attribute{
			"configured_status": schema.StringAttribute{
				MarkdownDescription: "Configured status of the network peering.",
				Default:             stringdefault.StaticString(string(upcloud.NetworkPeeringConfiguredStatusActive)),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(upcloud.NetworkPeeringConfiguredStatusActive),
						string(upcloud.NetworkPeeringConfiguredStatusDisabled)),
				},
			},
			"labels": utils.LabelsAttribute("network peering"),
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the network peering.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the network peering.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"network":      networkBlock("Local"),
			"peer_network": networkBlock("Peer"),
		},
	}
}

func setValues(ctx context.Context, data *networkPeeringModel, peering *upcloud.NetworkPeering) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	data.ConfiguredStatus = types.StringValue(string(peering.ConfiguredStatus))
	data.Name = types.StringValue(peering.Name)
	data.ID = types.StringValue(peering.UUID)

	data.Labels, respDiagnostics = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(peering.Labels))

	if len(data.Network) == 0 {
		data.Network = make([]networkModel, 1)
	}
	data.Network[0].UUID = types.StringValue(peering.Network.UUID)

	if len(data.PeerNetwork) == 0 {
		data.PeerNetwork = make([]networkModel, 1)
	}
	data.PeerNetwork[0].UUID = types.StringValue(peering.PeerNetwork.UUID)

	return respDiagnostics
}

func (r *networkPeeringResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data networkPeeringModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}

	apiReq := request.CreateNetworkPeeringRequest{
		ConfiguredStatus: upcloud.NetworkPeeringConfiguredStatus(data.ConfiguredStatus.ValueString()),
		Name:             data.Name.ValueString(),
		Labels:           utils.LabelsMapToSlice(labels),
		Network: request.NetworkPeeringNetwork{
			UUID: data.Network[0].UUID.ValueString(),
		},
		PeerNetwork: request.NetworkPeeringNetwork{
			UUID: data.PeerNetwork[0].UUID.ValueString(),
		},
	}

	peering, err := r.client.CreateNetworkPeering(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create network peering",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(waitForPeeringToLeaveProvisionedState(ctx, r.client, peering.UUID)...)
	resp.Diagnostics.Append(setValues(ctx, &data, peering)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *networkPeeringResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data networkPeeringModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	peering, err := r.client.GetNetworkPeering(ctx, &request.GetNetworkPeeringRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read network peering details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, peering)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *networkPeeringResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state networkPeeringModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var labels map[string]string
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	}
	labelsSlice := utils.LabelsMapToSlice(labels)

	apiReq := request.ModifyNetworkPeeringRequest{
		UUID: plan.ID.ValueString(),
		NetworkPeering: request.ModifyNetworkPeering{
			ConfiguredStatus: upcloud.NetworkPeeringConfiguredStatus(plan.ConfiguredStatus.ValueString()),
			Name:             plan.Name.ValueString(),
			Labels:           &labelsSlice,
		},
	}

	peering, err := r.client.ModifyNetworkPeering(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify network peering",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(waitForPeeringToLeaveProvisionedState(ctx, r.client, peering.UUID)...)
	resp.Diagnostics.Append(setValues(ctx, &plan, peering)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkPeeringResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data networkPeeringModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(waitForPeeringToLeaveProvisionedState(ctx, r.client, data.ID.ValueString())...)

	// Delete will fail with suitable error message if we get an error here.
	peering, _ := r.client.GetNetworkPeering(ctx, &request.GetNetworkPeeringRequest{
		UUID: data.ID.ValueString(),
	})
	if peering.State != upcloud.NetworkPeeringStateDisabled {
		_, err := r.client.ModifyNetworkPeering(ctx, &request.ModifyNetworkPeeringRequest{
			UUID: data.ID.ValueString(),
			NetworkPeering: request.ModifyNetworkPeering{
				ConfiguredStatus: upcloud.NetworkPeeringConfiguredStatusDisabled,
			},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to disable network peering",
				utils.ErrorDiagnosticDetail(err),
			)
		}

		_, err = r.client.WaitForNetworkPeeringState(ctx, &request.WaitForNetworkPeeringStateRequest{
			UUID:         data.ID.ValueString(),
			DesiredState: upcloud.NetworkPeeringStateDisabled,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to disable network peering",
				utils.ErrorDiagnosticDetail(err),
			)
		}
	}

	if err := r.client.DeleteNetworkPeering(ctx, &request.DeleteNetworkPeeringRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete network peering",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *networkPeeringResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitForPeeringToLeaveProvisionedState(ctx context.Context, svc *service.Service, uuid string) (diags diag.Diagnostics) {
	for {
		select {
		case <-ctx.Done():
			diags.AddError("Context cancelled", ctx.Err().Error())
			return
		default:
			peering, err := svc.GetNetworkPeering(ctx, &request.GetNetworkPeeringRequest{
				UUID: uuid,
			})
			if err != nil {
				diags.AddError("Unable to get network peering details", ctx.Err().Error())
				return
			}
			if peering.State != upcloud.NetworkPeeringStateProvisioning {
				return nil
			}

			tflog.Info(ctx, "waiting for peering to leave provisioned state", map[string]interface{}{"UUID": uuid})
		}
		time.Sleep(5 * time.Second)
	}
}
