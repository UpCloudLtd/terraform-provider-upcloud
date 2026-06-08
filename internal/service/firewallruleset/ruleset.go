package firewallruleset

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

type firewallRulesetModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	Stateful               types.Bool   `tfsdk:"stateful"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	DefaultDNSRulesEnabled types.Bool   `tfsdk:"default_dns_rules_enabled"`
	Labels                 types.Map    `tfsdk:"labels"`
	ServerUUID             types.String `tfsdk:"server_uuid"`
	Version                types.Int64  `tfsdk:"version"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
}

func (r *firewallRulesetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_ruleset"
}

func (r *firewallRulesetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

func (r *firewallRulesetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an UpCloud SDN firewall ruleset.",
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
			"stateful": schema.BoolAttribute{
				Description: "Whether rules are evaluated statefully. Create-only in API.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
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
		},
	}
}

func toAPILabels(ctx context.Context, labels types.Map) (*[]v9.FirewallRulesetCreateLabel, error) {
	if labels.IsNull() || labels.IsUnknown() {
		// Return pointer to empty slice instead of nil to explicitly clear labels
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
	if api.Stateful == nil {
		state.Stateful = types.BoolNull()
	} else {
		state.Stateful = types.BoolValue(*api.Stateful)
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

	body := v9.CreateFirewallRulesetJSONRequestBody{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueStringPointer()
	}
	if !plan.Stateful.IsNull() && !plan.Stateful.IsUnknown() {
		body.Stateful = plan.Stateful.ValueBoolPointer()
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
		if apiResp.HTTPResponse != nil && apiResp.Body != nil {
			detail = fmt.Sprintf("%s. Response: %s", detail, string(apiResp.Body))
		}
		resp.Diagnostics.AddError("Unable to create firewall ruleset", detail)
		return
	}

	if err := setRulesetValues(ctx, &plan, apiResp.JSON201); err != nil {
		resp.Diagnostics.AddError("Unable to create firewall ruleset", utils.ErrorDiagnosticDetail(err))
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

	// API uses PATCH semantics: omitted optional fields keep their existing values.
	// Always send: name (required) and labels (to allow clearing).
	// Conditionally send: all optional fields only when explicitly set in config.
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
