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

const (
	resourceDescrition = `Firewall rules are used to control network access of UpCloud servers. Each server has its own firewall rules and there should be only one ` + "`" + `upcloud_firewall_rules` + "`" + ` resource per server.
The firewall is enabled on public and utility network interfaces.`
	ruleDescription = `A single firewall rule. The rules are evaluated in order. The maximum number of firewall rules per server is 1000.

	Typical firewall rule should have ` + "`" + `action` + "`" + `, ` + "`" + `direction` + "`" + `, ` + "`" + `protocol` + "`" + `, ` + "`" + `family` + "`" + ` and at least one destination/source-address/port range.

	A default rule can be created by providing only ` + "`" + `action` + "`" + ` and ` + "`" + `direction` + "`" + ` attributes. Default rule should be defined last.

	If used, IP address and port ranges must have both start and end values specified. These can be the same value if only one IP address or port number is specified.
	Source and destination port numbers can only be set if the protocol is TCP or UDP.
	The ICMP type may only be set if the protocol is ICMP.`
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
	ID           types.String `tfsdk:"id"`
	ServerID     types.String `tfsdk:"server_id"`
	FirewallRule types.List   `tfsdk:"firewall_rule"`
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
		Description: resourceDescrition,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"server_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the server to be protected with the firewall rules.",
			},
		},
		Blocks: map[string]schema.Block{
			"firewall_rule": schema.ListNestedBlock{
				Description: ruleDescription,
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Description: "Action to take if the rule conditions are met. Valid values `accept | drop`.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("accept", "drop"),
							},
						},
						"comment": schema.StringAttribute{
							Description: "A comment for the rule.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(0, 250),
							},
						},
						"destination_address_start": schema.StringAttribute{
							Description: "The destination address range starts from this address",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsIPv4Address, validation.IsIPv6Address, validation.StringIsEmpty)),
							},
						},
						"destination_address_end": schema.StringAttribute{
							Description: "The destination address range ends from this address",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsIPv4Address, validation.IsIPv6Address, validation.StringIsEmpty)),
							},
						},
						"destination_port_start": schema.StringAttribute{
							Description: "The destination port range starts from this port number",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								portValidator{},
							},
						},
						"destination_port_end": schema.StringAttribute{
							Description: "The destination port range ends from this port number",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								portValidator{},
							},
						},
						"direction": schema.StringAttribute{
							Description: "The direction of network traffic this rule will be applied to. Valid values are `in` and `out`.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("in", "out"),
							},
						},
						"family": schema.StringAttribute{
							Description: "The address family of new firewall rule",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("IPv4", "IPv6"),
							},
						},
						"icmp_type": schema.StringAttribute{
							Description: "The ICMP type of the rule. Only valid if protocol is ICMP.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(0, 255),
							},
						},
						"protocol": schema.StringAttribute{
							Description: "The protocol of the rule. Possible values are `` (empty), `tcp`, `udp`, `icmp`.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("", "tcp", "udp", "icmp"),
							},
						},
						"source_address_start": schema.StringAttribute{
							Description: "The source address range starts from this address",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsIPv4Address, validation.IsIPv6Address, validation.StringIsEmpty)),
							},
						},
						"source_address_end": schema.StringAttribute{
							Description: "The source address range ends from this address",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								validatorutil.NewFrameworkStringValidator(validation.Any(validation.IsIPv4Address, validation.IsIPv6Address, validation.StringIsEmpty)),
							},
						},
						"source_port_start": schema.StringAttribute{
							Description: "The source port range starts from this port number",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								portValidator{},
							},
						},
						"source_port_end": schema.StringAttribute{
							Description: "The source port range ends from this port number",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								portValidator{},
							},
						},
					},
				},
			},
		},
	}
}

func setValues(ctx context.Context, data *firewallRulesModel, firewallRules *upcloud.FirewallRules) (diags diag.Diagnostics) {
	if firewallRules == nil || firewallRules.FirewallRules == nil {
		return nil
	}

	data.ID = types.StringValue(data.ServerID.ValueString())

	var dataFirewallRules []firewallRuleModel
	// data.FirewallRule is unknown when importing existing resource.
	if data.FirewallRule.IsNull() || data.FirewallRule.IsUnknown() {
		dataFirewallRules = make([]firewallRuleModel, len(firewallRules.FirewallRules))
	} else {
		diags.Append(data.FirewallRule.ElementsAs(ctx, &dataFirewallRules, false)...)
		if diags.HasError() {
			return diags
		}
	}

	for i, rule := range firewallRules.FirewallRules {
		if i >= len(dataFirewallRules) {
			dataFirewallRules = append(dataFirewallRules, firewallRuleModel{})
		}

		dataFirewallRules[i].Action = types.StringValue(rule.Action)
		dataFirewallRules[i].Comment = types.StringValue(rule.Comment)

		dataFirewallRules[i].DestinationAddressStart = types.StringValue(rule.DestinationAddressStart)
		dataFirewallRules[i].DestinationAddressEnd = types.StringValue(rule.DestinationAddressEnd)
		dataFirewallRules[i].DestinationPortStart = types.StringValue(rule.DestinationPortStart)
		dataFirewallRules[i].DestinationPortEnd = types.StringValue(rule.DestinationPortEnd)

		dataFirewallRules[i].Direction = types.StringValue(rule.Direction)
		dataFirewallRules[i].Family = types.StringValue(rule.Family)
		dataFirewallRules[i].ICMPType = types.StringValue(rule.ICMPType)
		dataFirewallRules[i].Protocol = types.StringValue(rule.Protocol)

		dataFirewallRules[i].SourceAddressStart = types.StringValue(rule.SourceAddressStart)
		dataFirewallRules[i].SourceAddressEnd = types.StringValue(rule.SourceAddressEnd)
		dataFirewallRules[i].SourcePortStart = types.StringValue(rule.SourcePortStart)
		dataFirewallRules[i].SourcePortEnd = types.StringValue(rule.SourcePortEnd)
	}

	data.FirewallRule, diags = types.ListValueFrom(ctx, data.FirewallRule.ElementType(ctx), dataFirewallRules)
	return
}

func buildFirewallRules(ctx context.Context, plan firewallRulesModel) ([]upcloud.FirewallRule, diag.Diagnostics) {
	var planFirewallRules []firewallRuleModel
	var diags diag.Diagnostics

	diags.Append(plan.FirewallRule.ElementsAs(ctx, &planFirewallRules, false)...)
	if diags.HasError() {
		return nil, diags
	}

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

	apiFirewallRules, diags := buildFirewallRules(ctx, data)
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

	if data.ID.ValueString() == "" && data.ServerID.ValueString() == "" {
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

	apiFirewallRules, diags := buildFirewallRules(ctx, data)
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
}

func (r *firewallRulesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("server_id"), req, resp)
}
