package firewallruleset

import (
	"context"
	"fmt"
	"net/http"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &firewallRulesetRuleResource{}
	_ resource.ResourceWithConfigure   = &firewallRulesetRuleResource{}
	_ resource.ResourceWithImportState = &firewallRulesetRuleResource{}
)

func NewFirewallRulesetRuleResource() resource.Resource {
	return &firewallRulesetRuleResource{}
}

type firewallRulesetRuleResource struct {
	client *v9.ClientWithResponses
}

type firewallRulesetRuleModel struct {
	ID                      types.String `tfsdk:"id"`
	RulesetUUID             types.String `tfsdk:"ruleset_uuid"`
	RuleID                  types.String `tfsdk:"rule_id"`
	Action                  types.String `tfsdk:"action"`
	Direction               types.String `tfsdk:"direction"`
	Family                  types.String `tfsdk:"family"`
	Protocol                types.String `tfsdk:"protocol"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	Comment                 types.String `tfsdk:"comment"`
	Position                types.Int64  `tfsdk:"position"`
	ICMPType                types.Int64  `tfsdk:"icmp_type"`
	SourceAddressCIDR       types.String `tfsdk:"source_address_cidr"`
	SourceAddressStart      types.String `tfsdk:"source_address_start"`
	SourceAddressEnd        types.String `tfsdk:"source_address_end"`
	SourcePortStart         types.Int64  `tfsdk:"source_port_start"`
	SourcePortEnd           types.Int64  `tfsdk:"source_port_end"`
	DestinationAddressCIDR  types.String `tfsdk:"destination_address_cidr"`
	DestinationAddressStart types.String `tfsdk:"destination_address_start"`
	DestinationAddressEnd   types.String `tfsdk:"destination_address_end"`
	DestinationPortStart    types.Int64  `tfsdk:"destination_port_start"`
	DestinationPortEnd      types.Int64  `tfsdk:"destination_port_end"`
}

func (r *firewallRulesetRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_ruleset_rule"
}

func (r *firewallRulesetRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

func (r *firewallRulesetRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents a single rule inside an UpCloud SDN firewall ruleset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Rule ID in {ruleset_uuid}/{rule_id} format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ruleset_uuid": schema.StringAttribute{
				Description: "Firewall ruleset UUID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rule_id": schema.StringAttribute{
				Description: "Firewall rule UUID.",
				Computed:    true,
			},
			"action": schema.StringAttribute{
				Description: "Rule action.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("accept", "drop"),
				},
			},
			"direction": schema.StringAttribute{
				Description: "Rule direction.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("in", "out"),
				},
			},
			"family": schema.StringAttribute{
				Description: "Address family.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("IPv4", "IPv6"),
				},
			},
			"protocol": schema.StringAttribute{
				Description: "IP protocol.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether rule is enabled.",
				Optional:    true,
			},
			"comment": schema.StringAttribute{
				Description: "Rule comment.",
				Optional:    true,
			},
			"position": schema.Int64Attribute{
				Description: "Rule order position.",
				Optional:    true,
			},
			"icmp_type": schema.Int64Attribute{
				Description: "ICMP type.",
				Optional:    true,
			},
			"source_address_cidr": schema.StringAttribute{
				Description: "Source CIDR.",
				Optional:    true,
			},
			"source_address_start": schema.StringAttribute{
				Description: "Source range start.",
				Optional:    true,
			},
			"source_address_end": schema.StringAttribute{
				Description: "Source range end.",
				Optional:    true,
			},
			"source_port_start": schema.Int64Attribute{
				Description: "Source port range start.",
				Optional:    true,
			},
			"source_port_end": schema.Int64Attribute{
				Description: "Source port range end.",
				Optional:    true,
			},
			"destination_address_cidr": schema.StringAttribute{
				Description: "Destination CIDR.",
				Optional:    true,
			},
			"destination_address_start": schema.StringAttribute{
				Description: "Destination range start.",
				Optional:    true,
			},
			"destination_address_end": schema.StringAttribute{
				Description: "Destination range end.",
				Optional:    true,
			},
			"destination_port_start": schema.Int64Attribute{
				Description: "Destination port range start.",
				Optional:    true,
			},
			"destination_port_end": schema.Int64Attribute{
				Description: "Destination port range end.",
				Optional:    true,
			},
		},
	}
}

func setRuleState(state *firewallRulesetRuleModel, api *v9.FirewallRulesetRuleDetailResponse) {
	if api.Uuid != nil {
		state.RuleID = types.StringValue(api.Uuid.String())
		state.ID = types.StringValue(utils.MarshalID(state.RulesetUUID.ValueString(), api.Uuid.String()))
	}

	state.Action = types.StringValue(interfaceString(api.Action))
	state.Direction = types.StringValue(interfaceString(api.Direction))
	state.Family = types.StringValue(interfaceString(api.Family))

	if api.Protocol == nil {
		state.Protocol = types.StringNull()
	} else {
		state.Protocol = types.StringValue(interfaceString(api.Protocol))
	}
	if api.Enabled == nil {
		state.Enabled = types.BoolNull()
	} else {
		state.Enabled = types.BoolValue(*api.Enabled)
	}
	if api.Comment == nil {
		state.Comment = types.StringNull()
	} else {
		state.Comment = types.StringValue(*api.Comment)
	}
	if api.Position == nil {
		state.Position = types.Int64Null()
	} else {
		state.Position = types.Int64Value(*api.Position)
	}
	if api.IcmpType == nil {
		state.ICMPType = types.Int64Null()
	} else {
		state.ICMPType = types.Int64Value(*api.IcmpType)
	}
	if api.SourceAddressCidr == nil {
		state.SourceAddressCIDR = types.StringNull()
	} else {
		state.SourceAddressCIDR = types.StringValue(*api.SourceAddressCidr)
	}
	if api.SourceAddressStart == nil {
		state.SourceAddressStart = types.StringNull()
	} else {
		state.SourceAddressStart = types.StringValue(*api.SourceAddressStart)
	}
	if api.SourceAddressEnd == nil {
		state.SourceAddressEnd = types.StringNull()
	} else {
		state.SourceAddressEnd = types.StringValue(*api.SourceAddressEnd)
	}
	if api.SourcePortStart == nil {
		state.SourcePortStart = types.Int64Null()
	} else {
		state.SourcePortStart = types.Int64Value(*api.SourcePortStart)
	}
	if api.SourcePortEnd == nil {
		state.SourcePortEnd = types.Int64Null()
	} else {
		state.SourcePortEnd = types.Int64Value(*api.SourcePortEnd)
	}
	if api.DestinationAddressCidr == nil {
		state.DestinationAddressCIDR = types.StringNull()
	} else {
		state.DestinationAddressCIDR = types.StringValue(*api.DestinationAddressCidr)
	}
	if api.DestinationAddressStart == nil {
		state.DestinationAddressStart = types.StringNull()
	} else {
		state.DestinationAddressStart = types.StringValue(*api.DestinationAddressStart)
	}
	if api.DestinationAddressEnd == nil {
		state.DestinationAddressEnd = types.StringNull()
	} else {
		state.DestinationAddressEnd = types.StringValue(*api.DestinationAddressEnd)
	}
	if api.DestinationPortStart == nil {
		state.DestinationPortStart = types.Int64Null()
	} else {
		state.DestinationPortStart = types.Int64Value(*api.DestinationPortStart)
	}
	if api.DestinationPortEnd == nil {
		state.DestinationPortEnd = types.Int64Null()
	} else {
		state.DestinationPortEnd = types.Int64Value(*api.DestinationPortEnd)
	}
}

func (r *firewallRulesetRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallRulesetRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := v9.CreateFirewallRulesetRuleJSONRequestBody{
		Action:    plan.Action.ValueString(),
		Direction: plan.Direction.ValueString(),
		Family:    plan.Family.ValueString(),
	}
	if !plan.Protocol.IsNull() && !plan.Protocol.IsUnknown() {
		body.Protocol = plan.Protocol.ValueString()
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		body.Enabled = plan.Enabled.ValueBoolPointer()
	}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		body.Comment = plan.Comment.ValueStringPointer()
	}
	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {
		body.Position = plan.Position.ValueInt64Pointer()
	}
	if !plan.ICMPType.IsNull() && !plan.ICMPType.IsUnknown() {
		body.IcmpType = plan.ICMPType.ValueInt64Pointer()
	}
	if !plan.SourceAddressCIDR.IsNull() && !plan.SourceAddressCIDR.IsUnknown() {
		body.SourceAddressCidr = plan.SourceAddressCIDR.ValueStringPointer()
	}
	if !plan.SourceAddressStart.IsNull() && !plan.SourceAddressStart.IsUnknown() {
		body.SourceAddressStart = plan.SourceAddressStart.ValueStringPointer()
	}
	if !plan.SourceAddressEnd.IsNull() && !plan.SourceAddressEnd.IsUnknown() {
		body.SourceAddressEnd = plan.SourceAddressEnd.ValueStringPointer()
	}
	if !plan.SourcePortStart.IsNull() && !plan.SourcePortStart.IsUnknown() {
		body.SourcePortStart = plan.SourcePortStart.ValueInt64Pointer()
	}
	if !plan.SourcePortEnd.IsNull() && !plan.SourcePortEnd.IsUnknown() {
		body.SourcePortEnd = plan.SourcePortEnd.ValueInt64Pointer()
	}
	if !plan.DestinationAddressCIDR.IsNull() && !plan.DestinationAddressCIDR.IsUnknown() {
		body.DestinationAddressCidr = plan.DestinationAddressCIDR.ValueStringPointer()
	}
	if !plan.DestinationAddressStart.IsNull() && !plan.DestinationAddressStart.IsUnknown() {
		body.DestinationAddressStart = plan.DestinationAddressStart.ValueStringPointer()
	}
	if !plan.DestinationAddressEnd.IsNull() && !plan.DestinationAddressEnd.IsUnknown() {
		body.DestinationAddressEnd = plan.DestinationAddressEnd.ValueStringPointer()
	}
	if !plan.DestinationPortStart.IsNull() && !plan.DestinationPortStart.IsUnknown() {
		body.DestinationPortStart = plan.DestinationPortStart.ValueInt64Pointer()
	}
	if !plan.DestinationPortEnd.IsNull() && !plan.DestinationPortEnd.IsUnknown() {
		body.DestinationPortEnd = plan.DestinationPortEnd.ValueInt64Pointer()
	}

	rulesetUUID, err := uuid.Parse(plan.RulesetUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	apiResp, err := r.client.CreateFirewallRulesetRuleWithResponse(ctx, rulesetUUID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create firewall ruleset rule", utils.ErrorDiagnosticDetail(err))
		return
	}
	if apiResp.StatusCode() != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Unable to create firewall ruleset rule",
			fmt.Sprintf("API returned unexpected status %s", apiResp.Status()),
		)
		return
	}

	setRuleState(&plan, apiResp.JSON201)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *firewallRulesetRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallRulesetRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	var rulesetUUIDStr, ruleIDStr string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(state.ID.ValueString(), &rulesetUUIDStr, &ruleIDStr)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.RulesetUUID = types.StringValue(rulesetUUIDStr)
	state.RuleID = types.StringValue(ruleIDStr)

	rulesetUUID, err := uuid.Parse(rulesetUUIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset UUID", utils.ErrorDiagnosticDetail(err))
		return
	}
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid rule ID", utils.ErrorDiagnosticDetail(err))
		return
	}

	apiResp, err := r.client.GetFirewallRulesetRuleWithResponse(ctx, rulesetUUID, ruleID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read firewall ruleset rule", utils.ErrorDiagnosticDetail(err))
		return
	}
	if apiResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to read firewall ruleset rule",
			fmt.Sprintf("API returned unexpected status %s", apiResp.Status()),
		)
		return
	}

	setRuleState(&state, apiResp.JSON200)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallRulesetRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state firewallRulesetRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rulesetUUIDStr, ruleIDStr string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(state.ID.ValueString(), &rulesetUUIDStr, &ruleIDStr)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.RulesetUUID = types.StringValue(rulesetUUIDStr)
	plan.RuleID = types.StringValue(ruleIDStr)
	plan.ID = state.ID

	body := v9.ModifyFirewallRulesetRuleJSONRequestBody{}
	if !plan.Action.IsNull() && !plan.Action.IsUnknown() {
		body.Action = plan.Action.ValueStringPointer()
	}
	if !plan.Direction.IsNull() && !plan.Direction.IsUnknown() {
		body.Direction = plan.Direction.ValueStringPointer()
	}
	if !plan.Family.IsNull() && !plan.Family.IsUnknown() {
		body.Family = plan.Family.ValueStringPointer()
	}
	if !plan.Protocol.IsNull() && !plan.Protocol.IsUnknown() {
		body.Protocol = plan.Protocol.ValueStringPointer()
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		body.Enabled = plan.Enabled.ValueBoolPointer()
	}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		body.Comment = plan.Comment.ValueStringPointer()
	}
	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {
		body.Position = plan.Position.ValueInt64Pointer()
	}
	if !plan.ICMPType.IsNull() && !plan.ICMPType.IsUnknown() {
		body.IcmpType = plan.ICMPType.ValueInt64Pointer()
	}
	if !plan.SourceAddressCIDR.IsNull() && !plan.SourceAddressCIDR.IsUnknown() {
		body.SourceAddressCidr = plan.SourceAddressCIDR.ValueStringPointer()
	}
	if !plan.SourceAddressStart.IsNull() && !plan.SourceAddressStart.IsUnknown() {
		body.SourceAddressStart = plan.SourceAddressStart.ValueStringPointer()
	}
	if !plan.SourceAddressEnd.IsNull() && !plan.SourceAddressEnd.IsUnknown() {
		body.SourceAddressEnd = plan.SourceAddressEnd.ValueStringPointer()
	}
	if !plan.SourcePortStart.IsNull() && !plan.SourcePortStart.IsUnknown() {
		body.SourcePortStart = plan.SourcePortStart.ValueInt64Pointer()
	}
	if !plan.SourcePortEnd.IsNull() && !plan.SourcePortEnd.IsUnknown() {
		body.SourcePortEnd = plan.SourcePortEnd.ValueInt64Pointer()
	}
	if !plan.DestinationAddressCIDR.IsNull() && !plan.DestinationAddressCIDR.IsUnknown() {
		body.DestinationAddressCidr = plan.DestinationAddressCIDR.ValueStringPointer()
	}
	if !plan.DestinationAddressStart.IsNull() && !plan.DestinationAddressStart.IsUnknown() {
		body.DestinationAddressStart = plan.DestinationAddressStart.ValueStringPointer()
	}
	if !plan.DestinationAddressEnd.IsNull() && !plan.DestinationAddressEnd.IsUnknown() {
		body.DestinationAddressEnd = plan.DestinationAddressEnd.ValueStringPointer()
	}
	if !plan.DestinationPortStart.IsNull() && !plan.DestinationPortStart.IsUnknown() {
		body.DestinationPortStart = plan.DestinationPortStart.ValueInt64Pointer()
	}
	if !plan.DestinationPortEnd.IsNull() && !plan.DestinationPortEnd.IsUnknown() {
		body.DestinationPortEnd = plan.DestinationPortEnd.ValueInt64Pointer()
	}

	rulesetUUID, err := uuid.Parse(rulesetUUIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset UUID", utils.ErrorDiagnosticDetail(err))
		return
	}
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid rule ID", utils.ErrorDiagnosticDetail(err))
		return
	}

	apiResp, err := r.client.ModifyFirewallRulesetRuleWithResponse(ctx, rulesetUUID, ruleID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update firewall ruleset rule", utils.ErrorDiagnosticDetail(err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to update firewall ruleset rule",
			fmt.Sprintf("API returned unexpected status %s", apiResp.Status()),
		)
		return
	}

	setRuleState(&plan, apiResp.JSON200)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *firewallRulesetRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state firewallRulesetRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rulesetUUIDStr, ruleIDStr string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(state.ID.ValueString(), &rulesetUUIDStr, &ruleIDStr)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rulesetUUID, err := uuid.Parse(rulesetUUIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset UUID", utils.ErrorDiagnosticDetail(err))
		return
	}
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid rule ID", utils.ErrorDiagnosticDetail(err))
		return
	}

	apiResp, err := r.client.DeleteFirewallRulesetRuleWithResponse(ctx, rulesetUUID, ruleID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete firewall ruleset rule", utils.ErrorDiagnosticDetail(err))
		return
	}
	if apiResp.StatusCode() != http.StatusNoContent && apiResp.StatusCode() != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Unable to delete firewall ruleset rule",
			fmt.Sprintf("API returned unexpected status %s", apiResp.Status()),
		)
	}
}

func (r *firewallRulesetRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var rulesetUUID, ruleID string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(req.ID, &rulesetUUID, &ruleID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ruleset_uuid"), rulesetUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule_id"), ruleID)...)
}
