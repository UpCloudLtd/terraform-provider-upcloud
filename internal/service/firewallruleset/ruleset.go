package firewallruleset

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &firewallRulesetResource{}
	_ resource.ResourceWithConfigure   = &firewallRulesetResource{}
	_ resource.ResourceWithImportState = &firewallRulesetResource{}
)

func NewFirewallRulesetResource() resource.Resource {
	return &firewallRulesetResource{}
}

type firewallRulesetResource struct {
	client *v9.ClientWithResponses
}

// ruleBlockModel is the inline rule model within a ruleset's rule list.
// Position is computed-only; it reflects the 1-based index assigned by the API.
type ruleBlockModel struct {
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

type firewallRulesetModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	DefaultDNSRulesEnabled types.Bool   `tfsdk:"default_dns_rules_enabled"`
	Labels                 types.Map    `tfsdk:"labels"`
	ServerUUID             types.String `tfsdk:"server_uuid"`
	Version                types.Int64  `tfsdk:"version"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	Rules                  types.List   `tfsdk:"rules"`
}

// ruleAttrTypes maps rule block attribute names to their Framework types.
// Must be kept in sync with ruleBlockModel tfsdk tags and the rules schema.
var ruleAttrTypes = map[string]attr.Type{
	"action":                    types.StringType,
	"direction":                 types.StringType,
	"family":                    types.StringType,
	"protocol":                  types.StringType,
	"enabled":                   types.BoolType,
	"comment":                   types.StringType,
	"position":                  types.Int64Type,
	"icmp_type":                 types.Int64Type,
	"source_address_cidr":       types.StringType,
	"source_address_start":      types.StringType,
	"source_address_end":        types.StringType,
	"source_port_start":         types.Int64Type,
	"source_port_end":           types.Int64Type,
	"destination_address_cidr":  types.StringType,
	"destination_address_start": types.StringType,
	"destination_address_end":   types.StringType,
	"destination_port_start":    types.Int64Type,
	"destination_port_end":      types.Int64Type,
}

// ruleObjectType returns the Framework object type for a single rule block element.
func ruleObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: ruleAttrTypes}
}

func (r *firewallRulesetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_ruleset"
}

func (r *firewallRulesetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

func (r *firewallRulesetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an UpCloud SDN firewall ruleset. Rules are managed as an ordered list; their position in the API is determined by their order in this list.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Firewall ruleset UUID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the firewall ruleset.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the firewall ruleset.",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the ruleset is enabled.",
				Optional:    true,
				Computed:    true,
			},
			"default_dns_rules_enabled": schema.BoolAttribute{
				Description: "Whether default DNS rules are enabled.",
				Optional:    true,
				Computed:    true,
			},
			"labels": schema.MapAttribute{
				Description: "Ruleset labels as key/value pairs.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"server_uuid": schema.StringAttribute{
				Description: "Optional server UUID to bind with this ruleset. Create-only in API.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.Int64Attribute{
				Description: "Ruleset version.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp.",
				Computed:    true,
			},
			"rules": schema.ListNestedAttribute{
				Description: "Ordered list of firewall rules. Rules are applied in list order; the position attribute reflects the 1-based index assigned by the API and cannot be configured.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Description: "Rule action.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("accept", "drop"),
							},
						},
						"direction": schema.StringAttribute{
							Description: "Traffic direction the rule applies to.",
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
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the rule is enabled.",
							Optional:    true,
							Computed:    true,
						},
						"comment": schema.StringAttribute{
							Description: "Rule comment.",
							Optional:    true,
							Computed:    true,
						},
						"position": schema.Int64Attribute{
							Description: "Rule position (1-based). Computed from the rule's index in the list; cannot be configured.",
							Computed:    true,
						},
						"icmp_type": schema.Int64Attribute{
							Description: "ICMP type number.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(0, 255),
							},
						},
						"source_address_cidr": schema.StringAttribute{
							Description: "Source CIDR block.",
							Optional:    true,
							Computed:    true,
						},
						"source_address_start": schema.StringAttribute{
							Description: "Start of source IP address range.",
							Optional:    true,
							Computed:    true,
						},
						"source_address_end": schema.StringAttribute{
							Description: "End of source IP address range.",
							Optional:    true,
							Computed:    true,
						},
						"source_port_start": schema.Int64Attribute{
							Description: "Start of source port range.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"source_port_end": schema.Int64Attribute{
							Description: "End of source port range.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"destination_address_cidr": schema.StringAttribute{
							Description: "Destination CIDR block.",
							Optional:    true,
							Computed:    true,
						},
						"destination_address_start": schema.StringAttribute{
							Description: "Start of destination IP address range.",
							Optional:    true,
							Computed:    true,
						},
						"destination_address_end": schema.StringAttribute{
							Description: "End of destination IP address range.",
							Optional:    true,
							Computed:    true,
						},
						"destination_port_start": schema.Int64Attribute{
							Description: "Start of destination port range.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"destination_port_end": schema.Int64Attribute{
							Description: "End of destination port range.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
					},
				},
			},
		},
	}
}

// ruleToCreate converts a ruleBlockModel to the API create body. Position is intentionally
// omitted; the API assigns positions sequentially based on the order of rules in the array.
func ruleToCreate(r ruleBlockModel) v9.FirewallRulesetFirewallRuleCreate {
	rule := v9.FirewallRulesetFirewallRuleCreate{
		Action:    r.Action.ValueString(),
		Direction: r.Direction.ValueString(),
		Family:    r.Family.ValueString(),
	}
	if !r.Protocol.IsNull() && !r.Protocol.IsUnknown() {
		rule.Protocol = r.Protocol.ValueString()
	}
	if !r.Enabled.IsNull() && !r.Enabled.IsUnknown() {
		rule.Enabled = r.Enabled.ValueBoolPointer()
	}
	if !r.Comment.IsNull() && !r.Comment.IsUnknown() {
		rule.Comment = r.Comment.ValueStringPointer()
	}
	if !r.ICMPType.IsNull() && !r.ICMPType.IsUnknown() {
		rule.IcmpType = r.ICMPType.ValueInt64Pointer()
	}
	if !r.SourceAddressCIDR.IsNull() && !r.SourceAddressCIDR.IsUnknown() {
		rule.SourceAddressCidr = r.SourceAddressCIDR.ValueStringPointer()
	}
	if !r.SourceAddressStart.IsNull() && !r.SourceAddressStart.IsUnknown() {
		rule.SourceAddressStart = r.SourceAddressStart.ValueStringPointer()
	}
	if !r.SourceAddressEnd.IsNull() && !r.SourceAddressEnd.IsUnknown() {
		rule.SourceAddressEnd = r.SourceAddressEnd.ValueStringPointer()
	}
	if !r.SourcePortStart.IsNull() && !r.SourcePortStart.IsUnknown() {
		rule.SourcePortStart = r.SourcePortStart.ValueInt64Pointer()
	}
	if !r.SourcePortEnd.IsNull() && !r.SourcePortEnd.IsUnknown() {
		rule.SourcePortEnd = r.SourcePortEnd.ValueInt64Pointer()
	}
	if !r.DestinationAddressCIDR.IsNull() && !r.DestinationAddressCIDR.IsUnknown() {
		rule.DestinationAddressCidr = r.DestinationAddressCIDR.ValueStringPointer()
	}
	if !r.DestinationAddressStart.IsNull() && !r.DestinationAddressStart.IsUnknown() {
		rule.DestinationAddressStart = r.DestinationAddressStart.ValueStringPointer()
	}
	if !r.DestinationAddressEnd.IsNull() && !r.DestinationAddressEnd.IsUnknown() {
		rule.DestinationAddressEnd = r.DestinationAddressEnd.ValueStringPointer()
	}
	if !r.DestinationPortStart.IsNull() && !r.DestinationPortStart.IsUnknown() {
		rule.DestinationPortStart = r.DestinationPortStart.ValueInt64Pointer()
	}
	if !r.DestinationPortEnd.IsNull() && !r.DestinationPortEnd.IsUnknown() {
		rule.DestinationPortEnd = r.DestinationPortEnd.ValueInt64Pointer()
	}
	return rule
}

// ruleFromAPI converts an API response rule into a ruleBlockModel.
func ruleFromAPI(api v9.FirewallRulesetRuleDetailResponse) ruleBlockModel {
	r := ruleBlockModel{
		Action:    types.StringValue(interfaceString(api.Action)),
		Direction: types.StringValue(interfaceString(api.Direction)),
		Family:    types.StringValue(interfaceString(api.Family)),
	}
	if api.Protocol == nil {
		r.Protocol = types.StringNull()
	} else {
		r.Protocol = types.StringValue(interfaceString(api.Protocol))
	}
	if api.Enabled == nil {
		r.Enabled = types.BoolNull()
	} else {
		r.Enabled = types.BoolValue(*api.Enabled)
	}
	if api.Comment == nil {
		r.Comment = types.StringNull()
	} else {
		r.Comment = types.StringValue(*api.Comment)
	}
	if api.Position == nil {
		r.Position = types.Int64Null()
	} else {
		r.Position = types.Int64Value(*api.Position)
	}
	if api.IcmpType == nil {
		r.ICMPType = types.Int64Null()
	} else {
		r.ICMPType = types.Int64Value(*api.IcmpType)
	}
	if api.SourceAddressCidr == nil || *api.SourceAddressCidr == "" {
		r.SourceAddressCIDR = types.StringNull()
	} else {
		r.SourceAddressCIDR = types.StringValue(*api.SourceAddressCidr)
	}
	if api.SourceAddressStart == nil || *api.SourceAddressStart == "" {
		r.SourceAddressStart = types.StringNull()
	} else {
		r.SourceAddressStart = types.StringValue(*api.SourceAddressStart)
	}
	if api.SourceAddressEnd == nil || *api.SourceAddressEnd == "" {
		r.SourceAddressEnd = types.StringNull()
	} else {
		r.SourceAddressEnd = types.StringValue(*api.SourceAddressEnd)
	}
	if api.SourcePortStart == nil {
		r.SourcePortStart = types.Int64Null()
	} else {
		r.SourcePortStart = types.Int64Value(*api.SourcePortStart)
	}
	if api.SourcePortEnd == nil {
		r.SourcePortEnd = types.Int64Null()
	} else {
		r.SourcePortEnd = types.Int64Value(*api.SourcePortEnd)
	}
	if api.DestinationAddressCidr == nil || *api.DestinationAddressCidr == "" {
		r.DestinationAddressCIDR = types.StringNull()
	} else {
		r.DestinationAddressCIDR = types.StringValue(*api.DestinationAddressCidr)
	}
	if api.DestinationAddressStart == nil || *api.DestinationAddressStart == "" {
		r.DestinationAddressStart = types.StringNull()
	} else {
		r.DestinationAddressStart = types.StringValue(*api.DestinationAddressStart)
	}
	if api.DestinationAddressEnd == nil || *api.DestinationAddressEnd == "" {
		r.DestinationAddressEnd = types.StringNull()
	} else {
		r.DestinationAddressEnd = types.StringValue(*api.DestinationAddressEnd)
	}
	if api.DestinationPortStart == nil {
		r.DestinationPortStart = types.Int64Null()
	} else {
		r.DestinationPortStart = types.Int64Value(*api.DestinationPortStart)
	}
	if api.DestinationPortEnd == nil {
		r.DestinationPortEnd = types.Int64Null()
	} else {
		r.DestinationPortEnd = types.Int64Value(*api.DestinationPortEnd)
	}
	return r
}

// putRules replaces the entire rule list for a ruleset using the bulk PUT endpoint.
// The API ignores any position field in the request body and assigns positions 1, 2, 3...
// based on the order of rules in the array.
func putRules(ctx context.Context, client *v9.ClientWithResponses, rulesetUUID uuid.UUID, rules []ruleBlockModel) ([]ruleBlockModel, error) {
	apiRules := make([]v9.FirewallRulesetFirewallRuleCreate, len(rules))
	for i, r := range rules {
		apiRules[i] = ruleToCreate(r)
	}

	body := v9.CreateMultipleFirewallRulesetRuleJSONRequestBody{
		Rules: apiRules,
	}

	resp, err := client.CreateMultipleFirewallRulesetRuleWithResponse(ctx, rulesetUUID, body)
	if err != nil {
		return nil, fmt.Errorf("calling bulk PUT rules: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("bulk PUT rules returned %s", resp.Status())
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("bulk PUT rules returned no response body")
	}

	result := make([]ruleBlockModel, len(resp.JSON200.Rules))
	for i, rule := range resp.JSON200.Rules {
		result[i] = ruleFromAPI(rule)
	}
	return result, nil
}

// readRules fetches the current rule list for a ruleset from the API.
func readRules(ctx context.Context, client *v9.ClientWithResponses, rulesetUUID uuid.UUID) ([]ruleBlockModel, error) {
	resp, err := client.ListFirewallRulesetRulesWithResponse(ctx, rulesetUUID, nil)
	if err != nil {
		return nil, fmt.Errorf("listing rules: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list rules returned %s", resp.Status())
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list rules returned no response body")
	}

	result := make([]ruleBlockModel, len(resp.JSON200.Rules))
	for i, rule := range resp.JSON200.Rules {
		result[i] = ruleFromAPI(rule)
	}
	return result, nil
}

// rulesFromAPI fetches current rules and converts them to a types.List.
func rulesFromAPI(ctx context.Context, client *v9.ClientWithResponses, rulesetUUID uuid.UUID) (types.List, error) {
	rules, err := readRules(ctx, client, rulesetUUID)
	if err != nil {
		return types.ListNull(ruleObjectType()), err
	}
	if len(rules) == 0 {
		return types.ListValueMust(ruleObjectType(), []attr.Value{}), nil
	}
	list, diags := types.ListValueFrom(ctx, ruleObjectType(), rules)
	if diags.HasError() {
		return types.ListNull(ruleObjectType()), fmt.Errorf("converting rules to list value")
	}
	return list, nil
}

// rulesFromPUT sends a bulk PUT and returns the resulting rules as a types.List.
func rulesFromPUT(ctx context.Context, client *v9.ClientWithResponses, rulesetUUID uuid.UUID, configRules []ruleBlockModel) (types.List, error) {
	rules, err := putRules(ctx, client, rulesetUUID, configRules)
	if err != nil {
		return types.ListNull(ruleObjectType()), err
	}
	if len(rules) == 0 {
		return types.ListValueMust(ruleObjectType(), []attr.Value{}), nil
	}
	list, diags := types.ListValueFrom(ctx, ruleObjectType(), rules)
	if diags.HasError() {
		return types.ListNull(ruleObjectType()), fmt.Errorf("converting PUT response rules to list value")
	}
	return list, nil
}

func toAPILabels(ctx context.Context, labels types.Map) (*[]v9.FirewallRulesetCreateLabel, error) {
	if labels.IsNull() || labels.IsUnknown() {
		emptyLabels := []v9.FirewallRulesetCreateLabel{}
		return &emptyLabels, nil
	}

	labelMap := map[string]string{}
	if diags := labels.ElementsAs(ctx, &labelMap, false); diags.HasError() {
		return nil, fmt.Errorf("unable to decode labels")
	}

	keys := make([]string, 0, len(labelMap))
	for k := range labelMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	apiLabels := make([]v9.FirewallRulesetCreateLabel, 0, len(labelMap))
	for _, k := range keys {
		apiLabels = append(apiLabels, v9.FirewallRulesetCreateLabel{Key: k, Value: labelMap[k]})
	}
	return &apiLabels, nil
}

func setRulesetValues(ctx context.Context, state *firewallRulesetModel, api *v9.FirewallRulesetDetailResponse) error {
	if api.Uuid != nil {
		state.ID = types.StringValue(api.Uuid.String())
	}
	if api.Name != nil {
		state.Name = types.StringValue(*api.Name)
	}
	if api.Description == nil {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(*api.Description)
	}
	if api.Enabled == nil {
		state.Enabled = types.BoolNull()
	} else {
		state.Enabled = types.BoolValue(*api.Enabled)
	}
	if api.DefaultDnsRulesEnabled == nil {
		state.DefaultDNSRulesEnabled = types.BoolNull()
	} else {
		state.DefaultDNSRulesEnabled = types.BoolValue(*api.DefaultDnsRulesEnabled)
	}
	if api.ServerUuid == nil {
		state.ServerUUID = types.StringNull()
	} else {
		state.ServerUUID = types.StringValue(api.ServerUuid.String())
	}
	if api.Version == nil {
		state.Version = types.Int64Null()
	} else {
		state.Version = types.Int64Value(int64(*api.Version))
	}
	if api.CreatedAt == nil {
		state.CreatedAt = types.StringNull()
	} else {
		state.CreatedAt = types.StringValue(api.CreatedAt.String())
	}
	if api.UpdatedAt == nil {
		state.UpdatedAt = types.StringNull()
	} else {
		state.UpdatedAt = types.StringValue(api.UpdatedAt.String())
	}
	if api.Labels == nil || len(*api.Labels) == 0 {
		state.Labels = types.MapNull(types.StringType)
	} else {
		labels := map[string]string{}
		for _, l := range *api.Labels {
			labels[l.Key] = l.Value
		}
		mapped, diags := types.MapValueFrom(ctx, types.StringType, labels)
		if diags.HasError() {
			return fmt.Errorf("unable to set labels")
		}
		state.Labels = mapped
	}
	return nil
}

func (r *firewallRulesetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallRulesetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labels, err := toAPILabels(ctx, plan.Labels)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create firewall ruleset", utils.ErrorDiagnosticDetail(err))
		return
	}

	stateful := true
	body := v9.CreateFirewallRulesetJSONRequestBody{
		Name:     plan.Name.ValueString(),
		Stateful: &stateful,
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueStringPointer()
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		body.Enabled = plan.Enabled.ValueBoolPointer()
	}
	if !plan.DefaultDNSRulesEnabled.IsNull() && !plan.DefaultDNSRulesEnabled.IsUnknown() {
		body.DefaultDnsRulesEnabled = plan.DefaultDNSRulesEnabled.ValueBoolPointer()
	}
	if labels != nil {
		body.Labels = labels
	}
	if !plan.ServerUUID.IsNull() && !plan.ServerUUID.IsUnknown() && plan.ServerUUID.ValueString() != "" {
		serverUUID, err := uuid.Parse(plan.ServerUUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid server UUID", utils.ErrorDiagnosticDetail(err))
			return
		}
		body.ServerUuid = &serverUUID
	}

	apiResp, err := r.client.CreateFirewallRulesetWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create firewall ruleset", utils.ErrorDiagnosticDetail(err))
		return
	}
	if apiResp.StatusCode() != http.StatusCreated {
		detail := fmt.Sprintf("API returned unexpected status %s", apiResp.Status())
		if apiResp.Body != nil {
			detail = fmt.Sprintf("%s. Response: %s", detail, string(apiResp.Body))
		}
		resp.Diagnostics.AddError("Unable to create firewall ruleset", detail)
		return
	}

	if err := setRulesetValues(ctx, &plan, apiResp.JSON201); err != nil {
		resp.Diagnostics.AddError("Unable to create firewall ruleset", utils.ErrorDiagnosticDetail(err))
		return
	}

	rulesetUUID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset UUID after create", utils.ErrorDiagnosticDetail(err))
		return
	}

	if !plan.Rules.IsNull() && !plan.Rules.IsUnknown() {
		var configRules []ruleBlockModel
		resp.Diagnostics.Append(plan.Rules.ElementsAs(ctx, &configRules, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Rules, err = rulesFromPUT(ctx, r.client, rulesetUUID, configRules)
		if err != nil {
			resp.Diagnostics.AddError("Unable to set firewall rules", utils.ErrorDiagnosticDetail(err))
			return
		}
	} else {
		plan.Rules, err = rulesFromAPI(ctx, r.client, rulesetUUID)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read firewall rules", utils.ErrorDiagnosticDetail(err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *firewallRulesetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallRulesetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	rulesetUUID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	apiResp, err := r.client.GetFirewallRulesetWithResponse(ctx, rulesetUUID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read firewall ruleset", utils.ErrorDiagnosticDetail(err))
		return
	}
	if apiResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to read firewall ruleset",
			fmt.Sprintf("API returned unexpected status %s", apiResp.Status()),
		)
		return
	}

	if err := setRulesetValues(ctx, &state, apiResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Unable to read firewall ruleset", utils.ErrorDiagnosticDetail(err))
		return
	}

	// Always refresh rules from API to keep state consistent.
	state.Rules, err = rulesFromAPI(ctx, r.client, rulesetUUID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read firewall rules", utils.ErrorDiagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallRulesetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state firewallRulesetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labels, err := toAPILabels(ctx, plan.Labels)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update firewall ruleset", utils.ErrorDiagnosticDetail(err))
		return
	}

	body := v9.ModifyFirewallRulesetJSONRequestBody{
		Name:   plan.Name.ValueStringPointer(),
		Labels: labels,
	}
	if !plan.Description.IsNull() {
		body.Description = plan.Description.ValueStringPointer()
	}
	if !plan.Enabled.IsNull() {
		body.Enabled = plan.Enabled.ValueBoolPointer()
	}
	if !plan.DefaultDNSRulesEnabled.IsNull() {
		body.DefaultDnsRulesEnabled = plan.DefaultDNSRulesEnabled.ValueBoolPointer()
	}

	rulesetUUID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	apiResp, err := r.client.ModifyFirewallRulesetWithResponse(ctx, rulesetUUID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update firewall ruleset", utils.ErrorDiagnosticDetail(err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to update firewall ruleset",
			fmt.Sprintf("API returned unexpected status %s", apiResp.Status()),
		)
		return
	}

	if err := setRulesetValues(ctx, &plan, apiResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Unable to update firewall ruleset", utils.ErrorDiagnosticDetail(err))
		return
	}

	if !plan.Rules.IsNull() && !plan.Rules.IsUnknown() {
		var configRules []ruleBlockModel
		resp.Diagnostics.Append(plan.Rules.ElementsAs(ctx, &configRules, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Rules, err = rulesFromPUT(ctx, r.client, rulesetUUID, configRules)
		if err != nil {
			resp.Diagnostics.AddError("Unable to update firewall rules", utils.ErrorDiagnosticDetail(err))
			return
		}
	} else {
		plan.Rules, err = rulesFromAPI(ctx, r.client, rulesetUUID)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read firewall rules", utils.ErrorDiagnosticDetail(err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *firewallRulesetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state firewallRulesetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rulesetUUID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	apiResp, err := r.client.DeleteFirewallRuleset(ctx, rulesetUUID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete firewall ruleset", utils.ErrorDiagnosticDetail(err))
		return
	}
	if apiResp.StatusCode != http.StatusNoContent && apiResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Unable to delete firewall ruleset",
			fmt.Sprintf("API returned unexpected status %s", apiResp.Status),
		)
	}
}

func (r *firewallRulesetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
