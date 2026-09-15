package firewallruleset

import (
	"testing"

	"github.com/google/uuid"
)

func TestParsePrivateFirewallRulesetRelationships_numericVersion(t *testing.T) {
	t.Parallel()

	rulesetUUID := uuid.MustParse("190f56d8-4b3f-4a89-9e5f-320fbc3d17c8")
	otherUUID := uuid.MustParse("0077fa3d-32db-4b09-9f5f-30d9e9afb565")

	// API may return version/applied_version as numbers while OpenAPI allows string|integer.
	body := []byte(`{
		"firewall_ruleset_relationships": {
			"private": [
				{
					"applied_version": 3,
					"created_at": "2026-08-25T10:15:30Z",
					"firewall_ruleset_uuid": "190f56d8-4b3f-4a89-9e5f-320fbc3d17c8",
					"last_applied_at": "2026-08-25T10:15:32Z",
					"server_uuid": "0077fa3d-32db-4b09-9f5f-30d9e9afb565",
					"type": "private",
					"updated_at": null,
					"version": 3
				}
			]
		}
	}`)

	parsed, err := parsePrivateFirewallRulesetRelationships(body)
	if err != nil {
		t.Fatalf("parsePrivateFirewallRulesetRelationships: %v", err)
	}

	if !privateFirewallRulesetAttached(parsed, rulesetUUID) {
		t.Fatalf("expected ruleset %s to be attached", rulesetUUID)
	}
	if privateFirewallRulesetAttached(parsed, otherUUID) {
		t.Fatalf("did not expect unrelated UUID %s to be attached", otherUUID)
	}
}

func TestParsePrivateFirewallRulesetRelationships_stringVersion(t *testing.T) {
	t.Parallel()

	rulesetUUID := uuid.MustParse("190f56d8-4b3f-4a89-9e5f-320fbc3d17c8")

	body := []byte(`{
		"firewall_ruleset_relationships": {
			"private": [
				{
					"applied_version": "3",
					"firewall_ruleset_uuid": "190f56d8-4b3f-4a89-9e5f-320fbc3d17c8",
					"type": "private",
					"version": "3"
				}
			]
		}
	}`)

	parsed, err := parsePrivateFirewallRulesetRelationships(body)
	if err != nil {
		t.Fatalf("parsePrivateFirewallRulesetRelationships: %v", err)
	}

	if !privateFirewallRulesetAttached(parsed, rulesetUUID) {
		t.Fatalf("expected ruleset %s to be attached", rulesetUUID)
	}
}
