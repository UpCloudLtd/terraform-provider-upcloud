package firewallruleset

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	serverUUIDDeprecationMessage = "server_uuid is no longer accepted when creating a firewall ruleset. Existing bindings are unchanged and remain read from the API. Use upcloud_server_private_firewall_ruleset for private SDN attachments, or upcloud_firewall_rules for classic public firewall rules."

	serverUUIDCreateSummary = "Cannot set server_uuid"
	serverUUIDCreateDetail  = "The API no longer accepts server_uuid when creating a firewall ruleset. Existing rulesets that already have a bound server are unchanged. For private SDN attachments, use upcloud_server_private_firewall_ruleset. For classic public firewall rules, use upcloud_firewall_rules."

	serverUUIDChangeSummary = "Cannot change server_uuid"
	serverUUIDChangeDetail  = "The API no longer supports binding a public firewall ruleset this way. Existing bindings are unchanged."
)

func serverUUIDConfigured(v types.String) bool {
	return !v.IsNull() && !v.IsUnknown() && v.ValueString() != ""
}

// serverUUIDPlanDiagnostic returns a summary and detail when server_uuid is set
// on create or changed on update. Empty strings mean the plan is allowed.
func serverUUIDPlanDiagnostic(creating bool, state, plan types.String) (summary, detail string) {
	if plan.IsUnknown() || !serverUUIDConfigured(plan) {
		return "", ""
	}
	if creating {
		return serverUUIDCreateSummary, serverUUIDCreateDetail
	}
	if state.IsUnknown() {
		return "", ""
	}
	if state.ValueString() == plan.ValueString() {
		return "", ""
	}
	return serverUUIDChangeSummary, serverUUIDChangeDetail
}

func interfaceString(v interface{}) string {
	if v == nil {
		return ""
	}

	s, ok := v.(string)
	if ok {
		return s
	}

	return fmt.Sprint(v)
}
