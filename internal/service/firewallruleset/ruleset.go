package firewallruleset

import (
	"context"
	"fmt"
	"maps"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                 = &firewallRulesetResource{}
	_ resource.ResourceWithConfigure    = &firewallRulesetResource{}
	_ resource.ResourceWithImportState  = &firewallRulesetResource{}
	_ resource.ResourceWithUpgradeState = &firewallRulesetResource{}
	_ resource.ResourceWithModifyPlan   = &firewallRulesetResource{}
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
	Version                types.Int64  `tfsdk:"version"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	Rules                  types.List   `tfsdk:"rules"`
}

type firewallRulesetModelV0 struct {
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
	resp.Schema = firewallRulesetSchema()
}

func firewallRulesetSchema() schema.Schema {
	s := firewallRulesetSchemaV0()
	attrs := maps.Clone(s.Attributes)
	delete(attrs, "server_uuid")
	s.Attributes = attrs
	s.Version = 1
	return s
}

func (r *firewallRulesetResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := firewallRulesetSchemaV0()
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior firewallRulesetModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, firewallRulesetModelFromV0(prior))...)
			},
		},
	}
}

func firewallRulesetModelFromV0(prior firewallRulesetModelV0) firewallRulesetModel {
	return firewallRulesetModel{
		ID:                     prior.ID,
		Name:                   prior.Name,
		Description:            prior.Description,
		Enabled:                prior.Enabled,
		DefaultDNSRulesEnabled: prior.DefaultDNSRulesEnabled,
		Labels:                 prior.Labels,
		Version:                prior.Version,
		CreatedAt:              prior.CreatedAt,
		UpdatedAt:              prior.UpdatedAt,
		Rules:                  prior.Rules,
	}
}

func firewallRulesetSchemaV0() schema.Schema {
	return schema.Schema{
		Description: "This resource represents an UpCloud SDN firewall ruleset. Rules are managed as an ordered list; their position in the API is determined by their order in the `rules` list.",
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
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the ruleset is enabled. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"default_dns_rules_enabled": schema.BoolAttribute{
				Description: "Whether default DNS rules are enabled. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"labels": utils.LabelsAttribute("firewall ruleset"),
			"server_uuid": schema.StringAttribute{
				Optional: true,
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
				Description: "Ordered list of firewall rules. Rules are applied in list order; the position attribute reflects the 1-based index assigned by the API and cannot be configured. If you leave out `rules`, Terraform deletes all rules in the ruleset, also rules added outside Terraform.",
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(ruleObjectType(), []attr.Value{})),
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
							Description: "IP protocol. Leave out or set `all` to match all protocols.",
							Optional:    true,
							Validators: []validator.String{
								// "" is not allowed
								stringvalidator.OneOf("tcp", "udp", "icmp", "esp", "ah", "all"),
							},
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the rule is enabled. Defaults to `true`.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
						},
						"comment": schema.StringAttribute{
							Description: "Rule comment.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"position": schema.Int64Attribute{
							Description: "Rule position (1-based). Computed from the rule's index in the list.",
							Computed:    true,
						},
						"icmp_type": schema.Int64Attribute{
							Description: "ICMP type number.",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(0, 255),
							},
						},
						"source_address_cidr": schema.StringAttribute{
							Description: "Source CIDR block.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"source_address_start": schema.StringAttribute{
							Description: "Start of source IP address range.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"source_address_end": schema.StringAttribute{
							Description: "End of source IP address range.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"source_port_start": schema.Int64Attribute{
							Description: "Start of source port range.",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"source_port_end": schema.Int64Attribute{
							Description: "End of source port range.",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"destination_address_cidr": schema.StringAttribute{
							Description: "Destination CIDR block.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"destination_address_start": schema.StringAttribute{
							Description: "Start of destination IP address range.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"destination_address_end": schema.StringAttribute{
							Description: "End of destination IP address range.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"destination_port_start": schema.Int64Attribute{
							Description: "Start of destination port range.",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"destination_port_end": schema.Int64Attribute{
							Description: "End of destination port range.",
							Optional:    true,
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
		Action:    v9.FirewallRulesetRuleAction(r.Action.ValueString()),
		Direction: v9.FirewallRulesetRuleDirection(r.Direction.ValueString()),
		Family:    v9.FirewallRulesetRuleFamily(r.Family.ValueString()),
		Enabled:   r.Enabled.ValueBoolPointer(),
		Comment:   r.Comment.ValueStringPointer(),
		IcmpType:  r.ICMPType.ValueInt64Pointer(),

		SourceAddressCidr:       r.SourceAddressCIDR.ValueStringPointer(),
		SourceAddressStart:      r.SourceAddressStart.ValueStringPointer(),
		SourceAddressEnd:        r.SourceAddressEnd.ValueStringPointer(),
		SourcePortStart:         r.SourcePortStart.ValueInt64Pointer(),
		SourcePortEnd:           r.SourcePortEnd.ValueInt64Pointer(),
		DestinationAddressCidr:  r.DestinationAddressCIDR.ValueStringPointer(),
		DestinationAddressStart: r.DestinationAddressStart.ValueStringPointer(),
		DestinationAddressEnd:   r.DestinationAddressEnd.ValueStringPointer(),
		DestinationPortStart:    r.DestinationPortStart.ValueInt64Pointer(),
		DestinationPortEnd:      r.DestinationPortEnd.ValueInt64Pointer(),
	}
	if !r.Protocol.IsNull() {
		p := v9.FirewallRulesetRuleProtocol(r.Protocol.ValueString())
		rule.Protocol = &p
	}
	return rule
}

// unsetString maps an empty or missing API string to null.
func unsetString(v *string) types.String {
	if v == nil || *v == "" {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

// unsetInt64 maps the API's unset sentinel (-1 for icmp_type, 0 for ports) to null.
func unsetInt64(v *int64, unset int64) types.Int64 {
	if v == nil || *v == unset {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}

// ruleFromAPI converts an API response rule into a ruleBlockModel.
func ruleFromAPI(api v9.FirewallRulesetRuleDetailResponse) ruleBlockModel {
	r := ruleBlockModel{
		Action:    types.StringValue(interfaceString(api.Action)),
		Direction: types.StringValue(interfaceString(api.Direction)),
		Family:    types.StringValue(interfaceString(api.Family)),
		Enabled:   types.BoolPointerValue(api.Enabled),
		Comment:   unsetString(api.Comment),
		Position:  types.Int64PointerValue(api.Position),
		ICMPType:  unsetInt64(api.IcmpType, -1),

		SourceAddressCIDR:       unsetString(api.SourceAddressCidr),
		SourceAddressStart:      unsetString(api.SourceAddressStart),
		SourceAddressEnd:        unsetString(api.SourceAddressEnd),
		DestinationAddressCIDR:  unsetString(api.DestinationAddressCidr),
		DestinationAddressStart: unsetString(api.DestinationAddressStart),
		DestinationAddressEnd:   unsetString(api.DestinationAddressEnd),

		SourcePortStart:      unsetInt64(api.SourcePortStart, 0),
		SourcePortEnd:        unsetInt64(api.SourcePortEnd, 0),
		DestinationPortStart: unsetInt64(api.DestinationPortStart, 0),
		DestinationPortEnd:   unsetInt64(api.DestinationPortEnd, 0),
	}
	// The API might return "all" or "" for a rule without a protocol.
	if p := interfaceString(api.Protocol); p == "" || p == "all" {
		r.Protocol = types.StringNull()
	} else {
		r.Protocol = types.StringValue(p)
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

// rulesToList converts rules to a types.List. No rules give an empty list, never null.
func rulesToList(ctx context.Context, rules []ruleBlockModel) (types.List, error) {
	if len(rules) == 0 {
		return types.ListValueMust(ruleObjectType(), []attr.Value{}), nil
	}
	list, diags := types.ListValueFrom(ctx, ruleObjectType(), rules)
	if diags.HasError() {
		return types.ListNull(ruleObjectType()), fmt.Errorf("converting rules to list value")
	}
	return list, nil
}

// keepProtocolAll keeps protocol "all" from the prior rule at the same index.
// The API might return a rule without a protocol and a rule with "all" the same way, so ruleFromAPI reads both as null.
func keepProtocolAll(rules, prior []ruleBlockModel) {
	for i := range rules {
		if i < len(prior) && rules[i].Protocol.IsNull() && prior[i].Protocol.ValueString() == "all" {
			rules[i].Protocol = prior[i].Protocol
		}
	}
}

// rulesFromAPI fetches current rules.
func rulesFromAPI(ctx context.Context, client *v9.ClientWithResponses, rulesetUUID uuid.UUID, prior types.List) (types.List, error) {
	rules, err := readRules(ctx, client, rulesetUUID)
	if err != nil {
		return types.ListNull(ruleObjectType()), err
	}
	var priorRules []ruleBlockModel
	if !prior.IsNull() && !prior.IsUnknown() {
		if diags := prior.ElementsAs(ctx, &priorRules, false); diags.HasError() {
			return types.ListNull(ruleObjectType()), fmt.Errorf("reading prior rules")
		}
	}
	keepProtocolAll(rules, priorRules)
	return rulesToList(ctx, rules)
}

// putPlannedRules replaces all rules with the planned rules. An empty plan clears all rules.
func putPlannedRules(ctx context.Context, client *v9.ClientWithResponses, rulesetUUID uuid.UUID, planned types.List) (types.List, error) {
	var configRules []ruleBlockModel
	if diags := planned.ElementsAs(ctx, &configRules, false); diags.HasError() {
		return types.ListNull(ruleObjectType()), fmt.Errorf("reading planned rules")
	}
	rules, err := putRules(ctx, client, rulesetUUID, configRules)
	if err != nil {
		return types.ListNull(ruleObjectType()), err
	}
	keepProtocolAll(rules, configRules)
	return rulesToList(ctx, rules)
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
	state.Description = unsetString(api.Description)
	state.Enabled = types.BoolPointerValue(api.Enabled)
	state.DefaultDNSRulesEnabled = types.BoolPointerValue(api.DefaultDnsRulesEnabled)
	if api.Version == nil {
		state.Version = types.Int64Null()
	} else {
		state.Version = types.Int64Value(*api.Version)
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
	labels := map[string]string{}
	if api.Labels != nil {
		for _, l := range *api.Labels {
			labels[l.Key] = l.Value
		}
	}
	mapped, diags := types.MapValueFrom(ctx, types.StringType, labels)
	if diags.HasError() {
		return fmt.Errorf("unable to set labels")
	}
	state.Labels = mapped
	return nil
}

// ModifyPlan marks the fields that the API updates as unknown when the plan changes the ruleset.
// The framework does this only for changes it sees before plan modifiers run. A change made by the labels plan modifier needs it here.
func (r *firewallRulesetResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() || req.Plan.Raw.Equal(req.State.Raw) {
		return
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("updated_at"), types.StringUnknown())...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("version"), types.Int64Unknown())...)
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
		Name:                   plan.Name.ValueString(),
		Stateful:               &stateful,
		Enabled:                plan.Enabled.ValueBoolPointer(),
		DefaultDnsRulesEnabled: plan.DefaultDNSRulesEnabled.ValueBoolPointer(),
		Description:            plan.Description.ValueStringPointer(),
	}
	if labels != nil {
		body.Labels = labels
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

	created := apiResp.JSON201
	// The create endpoint might ignore enabled=false, so disable the ruleset with a PATCH.
	if !plan.Enabled.ValueBool() && created != nil && created.Uuid != nil && created.Enabled != nil && *created.Enabled {
		disabled := false
		patchResp, err := r.client.ModifyFirewallRulesetWithResponse(ctx, *created.Uuid, v9.ModifyFirewallRulesetJSONRequestBody{Enabled: &disabled})
		if err != nil {
			resp.Diagnostics.AddError("Unable to disable firewall ruleset after create", utils.ErrorDiagnosticDetail(err))
			return
		}
		if patchResp.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError("Unable to disable firewall ruleset after create", fmt.Sprintf("API returned unexpected status %s", patchResp.Status()))
			return
		}
		created = patchResp.JSON200
	}

	if created == nil {
		resp.Diagnostics.AddError("Unable to create firewall ruleset", "API returned no ruleset data")
		return
	}

	if err := setRulesetValues(ctx, &plan, created); err != nil {
		resp.Diagnostics.AddError("Unable to create firewall ruleset", utils.ErrorDiagnosticDetail(err))
		return
	}

	rulesetUUID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset UUID after create", utils.ErrorDiagnosticDetail(err))
		return
	}

	plan.Rules, err = putPlannedRules(ctx, r.client, rulesetUUID, plan.Rules)
	if err != nil {
		resp.Diagnostics.AddError("Unable to set firewall rules", utils.ErrorDiagnosticDetail(err))
		return
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
	state.Rules, err = rulesFromAPI(ctx, r.client, rulesetUUID, state.Rules)
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
		Name:                   plan.Name.ValueStringPointer(),
		Labels:                 labels,
		Enabled:                plan.Enabled.ValueBoolPointer(),
		DefaultDnsRulesEnabled: plan.DefaultDNSRulesEnabled.ValueBoolPointer(),
	}
	// Send an empty string to clear the description. The API ignores a missing field.
	description := plan.Description.ValueString()
	body.Description = &description

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

	plan.Rules, err = putPlannedRules(ctx, r.client, rulesetUUID, plan.Rules)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update firewall rules", utils.ErrorDiagnosticDetail(err))
		return
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
	if apiResp.Body != nil {
		defer func() {
			if err := apiResp.Body.Close(); err != nil {
				resp.Diagnostics.AddError("Unable to delete firewall ruleset", fmt.Sprintf("unable to close API response body: %s", err))
			}
		}()
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
