package firewallruleset

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// privateFirewallRulesetRelationshipsResponse is a minimal list-response shape used
// instead of the generated v9 model. The model unmarshaling fails when the API returns
// numbers. Omit those fields so encoding/json ignores them.
type privateFirewallRulesetRelationshipsResponse struct {
	FirewallRulesetRelationships struct {
		Private []struct {
			FirewallRulesetUUID *uuid.UUID `json:"firewall_ruleset_uuid"`
		} `json:"private"`
	} `json:"firewall_ruleset_relationships"`
}

func parsePrivateFirewallRulesetRelationships(body []byte) (privateFirewallRulesetRelationshipsResponse, error) {
	var parsed privateFirewallRulesetRelationshipsResponse
	err := json.Unmarshal(body, &parsed)
	return parsed, err
}

func privateFirewallRulesetAttached(parsed privateFirewallRulesetRelationshipsResponse, rulesetUUID uuid.UUID) bool {
	for _, rel := range parsed.FirewallRulesetRelationships.Private {
		if rel.FirewallRulesetUUID != nil && *rel.FirewallRulesetUUID == rulesetUUID {
			return true
		}
	}
	return false
}

var (
	_ resource.Resource                = &serverPrivateFirewallRulesetResource{}
	_ resource.ResourceWithConfigure   = &serverPrivateFirewallRulesetResource{}
	_ resource.ResourceWithImportState = &serverPrivateFirewallRulesetResource{}
)

func NewServerPrivateFirewallRulesetResource() resource.Resource {
	return &serverPrivateFirewallRulesetResource{}
}

type serverPrivateFirewallRulesetResource struct {
	client *v9.ClientWithResponses
}

type serverPrivateFirewallRulesetModel struct {
	ID        types.String `tfsdk:"id"`
	ServerID  types.String `tfsdk:"server_id"`
	RulesetID types.String `tfsdk:"ruleset_id"`
}

func (r *serverPrivateFirewallRulesetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_private_firewall_ruleset"
}

func (r *serverPrivateFirewallRulesetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

func (r *serverPrivateFirewallRulesetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an attachment of an UpCloud SDN private firewall ruleset to a server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in format `{server_id}/{ruleset_id}`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Description: "The UUID of the server to attach the firewall ruleset to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ruleset_id": schema.StringAttribute{
				Description: "The UUID of the firewall ruleset to attach.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *serverPrivateFirewallRulesetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverPrivateFirewallRulesetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverUUID, err := uuid.Parse(plan.ServerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid server UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	rulesetUUID, err := uuid.Parse(plan.RulesetID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	var body v9.AttachPrivateFirewallRulesetJSONRequestBody
	body.FirewallRuleset.FirewallRulesetUuid = rulesetUUID

	apiResp, err := r.client.AttachPrivateFirewallRulesetWithResponse(ctx, serverUUID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to attach private firewall ruleset", utils.ErrorDiagnosticDetail(err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		detail := fmt.Sprintf("API returned unexpected status %s", apiResp.Status())
		if apiResp.Body != nil {
			detail = fmt.Sprintf("%s. Response: %s", detail, string(apiResp.Body))
		}
		resp.Diagnostics.AddError("Unable to attach private firewall ruleset", detail)
		return
	}

	plan.ID = types.StringValue(utils.MarshalID(plan.ServerID.ValueString(), plan.RulesetID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serverPrivateFirewallRulesetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverPrivateFirewallRulesetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	var serverID, rulesetID string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(state.ID.ValueString(), &serverID, &rulesetID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid server UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	rulesetUUID, err := uuid.Parse(rulesetID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	httpResp, err := r.client.ListPrivateFirewallRulesetRelationships(ctx, serverUUID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read private firewall ruleset relationships", utils.ErrorDiagnosticDetail(err))
		return
	}
	defer func() { _ = httpResp.Body.Close() }()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read private firewall ruleset relationships", utils.ErrorDiagnosticDetail(err))
		return
	}

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		detail := fmt.Sprintf("API returned unexpected status %s", httpResp.Status)
		if len(body) > 0 {
			detail = fmt.Sprintf("%s. Response: %s", detail, string(body))
		}
		resp.Diagnostics.AddError("Unable to read private firewall ruleset relationships", detail)
		return
	}

	parsed, err := parsePrivateFirewallRulesetRelationships(body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read private firewall ruleset relationships", utils.ErrorDiagnosticDetail(err))
		return
	}

	if !privateFirewallRulesetAttached(parsed, rulesetUUID) {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ServerID = types.StringValue(serverID)
	state.RulesetID = types.StringValue(rulesetID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serverPrivateFirewallRulesetResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "All attributes of server private firewall ruleset require replacement")
}

func (r *serverPrivateFirewallRulesetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverPrivateFirewallRulesetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var serverID, rulesetID string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(state.ID.ValueString(), &serverID, &rulesetID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid server UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	rulesetUUID, err := uuid.Parse(rulesetID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	var body v9.DetachPrivateFirewallRulesetJSONRequestBody
	body.FirewallRuleset.FirewallRulesetUuid = rulesetUUID

	apiResp, err := r.client.DetachPrivateFirewallRulesetWithResponse(ctx, serverUUID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to detach private firewall ruleset", utils.ErrorDiagnosticDetail(err))
		return
	}
	if apiResp.StatusCode() == http.StatusNotFound {
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		detail := fmt.Sprintf("API returned unexpected status %s", apiResp.Status())
		if apiResp.Body != nil {
			detail = fmt.Sprintf("%s. Response: %s", detail, string(apiResp.Body))
		}
		resp.Diagnostics.AddError("Unable to detach private firewall ruleset", detail)
		return
	}
}

func (r *serverPrivateFirewallRulesetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
