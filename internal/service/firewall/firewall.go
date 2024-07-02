package firewall

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	validatorutil "github.com/UpCloudLtd/terraform-provider-upcloud/internal/validator"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	_ resource.Resource                = &firewallRulesResource{}
	_ resource.ResourceWithConfigure   = &firewallRulesResource{}
	_ resource.ResourceWithImportState = &firewallRulesResource{}
)

func NewFirewallRulesResource() resource.Resource {
	return &firewallRulesResource{}
}

type firewallRulesResource struct {
	client *service.Service
}

func (r *firewallRulesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_rules"
}

// Configure adds the provider configured client to the resource.
func (r *firewallRulesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type firewallRulesModel struct {
	ID           types.String        `tfsdk:"id"`
	ServerID     types.String        `tfsdk:"server_id"`
	FirewallRule []firewallRuleModel `tfsdk:"firewall_rule"`
}

type firewallRuleModel struct {
	Action                  types.String `tfsdk:"action"`
	Comment                 types.String `tfsdk:"comment"`
	DestinationAddressStart types.String `tfsdk:"destination_address_start"`
	DestinationAddressEnd   types.String `tfsdk:"destination_address_end"`
	DestinationPortStart    types.String `tfsdk:"destination_port_start"`
	DestinationPortEnd      types.String `tfsdk:"destination_port_end"`
	Direction               types.String `tfsdk:"direction"`
	Family                  types.String `tfsdk:"family"`
	ICMPType                types.String `tfsdk:"icmp_type"`
	Protocol                types.String `tfsdk:"protocol"`
	SourceAddressStart      types.String `tfsdk:"source_address_start"`
	SourceAddressEnd        types.String `tfsdk:"source_address_end"`
	SourcePortStart         types.String `tfsdk:"source_port_start"`
	SourcePortEnd           types.String `tfsdk:"source_port_end"`
}

func (r *firewallRulesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "FirewallRuless can be used to connect multiple Private Networks. UpCloud Servers on any attached network can communicate directly with each other.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"server_id": schema.StringAttribute{
				Required:    true,
				Description: "The unique id of the server to be protected the firewall rules",
			},
			"firewall_rule": schema.ListNestedAttribute{
				Description: `A single firewall rule.
				If used, IP address and port ranges must have both start and end values specified. These can be the same value if only one IP address or port number is specified.
				Source and destination port numbers can only be set if the protocol is TCP or UDP.
				The ICMP type may only be set if the protocol is ICMP.
				Typical firewall rule should have "action", "direction", "protocol", "family" and at least one destination/source-address/port range.
				The default rule can be created by providing only "action" and "direction" attributes. Default rule should be defined last.`,
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Description: `The action to take on the packet. Possible values are "accept" and "drop".`,
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("accept", "drop"),
							},
						},
						"comment": schema.StringAttribute{
							Description: "A comment for the rule.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(0, 250),
							},
						},
						"destination_address_start": schema.StringAttribute{
							Description: "The destination address range starts from this address",
							Optional:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsIPv4Address, validation.IsIPv6Address, validation.StringIsEmpty)),
							},
						},
						"destination_address_end": schema.StringAttribute{
							Description: "The destination address range ends from this address",
							Optional:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsIPv4Address, validation.IsIPv6Address, validation.StringIsEmpty)),
							},
						},
						"destination_port_start": schema.StringAttribute{
							Description: "The destination port range starts from this port number",
							Optional:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsPortNumber, validation.StringIsEmpty)),
							},
						},
						"destination_port_end": schema.StringAttribute{
							Description: "The destination port range ends from this port number",
							Optional:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsPortNumber, validation.StringIsEmpty)),
							},
						},
						"direction": schema.StringAttribute{
							Description: "The direction of network traffic this rule will be applied to. Possible values are `in` and `out`.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("in", "out"),
							},
						},
						"family": schema.StringAttribute{
							Description: "The address family of new firewall rule",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("IPv4", "IPv6"),
							},
						},
						"icmp_type": schema.StringAttribute{
							Description: "The ICMP type of the rule. Only valid if protocol is ICMP.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(0, 255),
							},
						},
						"protocol": schema.StringAttribute{
							Description: "The protocol of the rule. Possible values are `` (empty), `tcp`, `udp`, `icmp`.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("", "tcp", "udp", "icmp"),
							},
						},
						"source_address_start": schema.StringAttribute{
							Description: "The source address range starts from this address",
							Optional:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsIPv4Address, validation.IsIPv6Address, validation.StringIsEmpty)),
							},
						},
						"source_address_end": schema.StringAttribute{
							Description: "The source address range ends from this address",
							Optional:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsIPv4Address, validation.IsIPv6Address, validation.StringIsEmpty)),
							},
						},
						"source_port_start": schema.StringAttribute{
							Description: "The source port range starts from this port number",
							Optional:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsPortNumber, validation.StringIsEmpty)),
							},
						},
						"source_port_end": schema.StringAttribute{
							Description: "The source port range ends from this port number",
							Optional:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsPortNumber, validation.StringIsEmpty)),
							},
						},
					},
				},
			},
		},
	}
}

func setValues(_ context.Context, data *firewallRulesModel, firewallRules *upcloud.FirewallRules) diag.Diagnostics {
	if firewallRules == nil || firewallRules.FirewallRules == nil {
		return nil
	}

	data.FirewallRule = make([]firewallRuleModel, len(firewallRules.FirewallRules))
	for i, rule := range firewallRules.FirewallRules {
		data.FirewallRule[i].Action = types.StringValue(rule.Action)
		data.FirewallRule[i].Comment = types.StringValue(rule.Comment)
		data.FirewallRule[i].DestinationAddressStart = types.StringValue(rule.DestinationAddressStart)
		data.FirewallRule[i].DestinationAddressEnd = types.StringValue(rule.DestinationAddressEnd)
		data.FirewallRule[i].DestinationPortStart = types.StringValue(rule.DestinationPortStart)
		data.FirewallRule[i].DestinationPortEnd = types.StringValue(rule.DestinationPortEnd)
		data.FirewallRule[i].Direction = types.StringValue(rule.Direction)
		data.FirewallRule[i].Family = types.StringValue(rule.Family)
		data.FirewallRule[i].ICMPType = types.StringValue(rule.ICMPType)
		data.FirewallRule[i].Protocol = types.StringValue(rule.Protocol)
		data.FirewallRule[i].SourceAddressStart = types.StringValue(rule.SourceAddressStart)
		data.FirewallRule[i].SourceAddressEnd = types.StringValue(rule.SourceAddressEnd)
		data.FirewallRule[i].SourcePortStart = types.StringValue(rule.SourcePortStart)
		data.FirewallRule[i].SourcePortEnd = types.StringValue(rule.SourcePortEnd)
	}

	return nil
}

func buildFirewallRules(_ context.Context, _ []firewallRuleModel) ([]upcloud.FirewallRule, diag.Diagnostics) {
	var planFirewallRules []firewallRuleModel

	firewallRules := make([]upcloud.FirewallRule, 0)

	for _, rule := range planFirewallRules {
		firewallRules = append(firewallRules, upcloud.FirewallRule{
			Action:                  rule.Action.ValueString(),
			Comment:                 rule.Comment.ValueString(),
			DestinationAddressStart: rule.DestinationAddressStart.ValueString(),
			DestinationAddressEnd:   rule.DestinationAddressEnd.ValueString(),
			DestinationPortStart:    rule.DestinationPortStart.ValueString(),
			DestinationPortEnd:      rule.DestinationPortEnd.ValueString(),
			Direction:               rule.Direction.ValueString(),
			Family:                  rule.Family.ValueString(),
			ICMPType:                rule.ICMPType.ValueString(),
			Protocol:                rule.Protocol.ValueString(),
			SourceAddressStart:      rule.SourceAddressStart.ValueString(),
			SourceAddressEnd:        rule.SourceAddressEnd.ValueString(),
			SourcePortStart:         rule.SourcePortStart.ValueString(),
			SourcePortEnd:           rule.SourcePortEnd.ValueString(),
		})
	}

	return firewallRules, nil
}

func (r *firewallRulesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data firewallRulesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiFirewallRules, diags := buildFirewallRules(ctx, data.FirewallRule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := request.CreateFirewallRulesRequest{
		ServerUUID:    data.ServerID.ValueString(),
		FirewallRules: apiFirewallRules,
	}

	err := r.client.CreateFirewallRules(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create firewall rules",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, &upcloud.FirewallRules{FirewallRules: apiFirewallRules})...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *firewallRulesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data firewallRulesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	firewallRules, err := r.client.GetFirewallRules(ctx, &request.GetFirewallRulesRequest{
		ServerUUID: data.ServerID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read firewall rules",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, firewallRules)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *firewallRulesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data firewallRulesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiFirewallRules, diags := buildFirewallRules(ctx, data.FirewallRule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := request.CreateFirewallRulesRequest{
		ServerUUID:    data.ServerID.ValueString(),
		FirewallRules: apiFirewallRules,
	}

	err := r.client.CreateFirewallRules(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create firewall rules",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, &upcloud.FirewallRules{FirewallRules: apiFirewallRules})...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *firewallRulesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data firewallRulesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := request.CreateFirewallRulesRequest{
		ServerUUID:    data.ServerID.ValueString(),
		FirewallRules: nil,
	}

	err := r.client.CreateFirewallRules(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete firewall rules",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, &upcloud.FirewallRules{FirewallRules: nil})...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *firewallRulesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("server_id"), req, resp)
}
