package firewallruleset

import (
	"testing"

	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRuleFromAPIMapsUnsetSentinelsToNull(t *testing.T) {
	t.Parallel()

	unsetIcmp := int64(-1)
	unsetPort := int64(0)
	zeroIcmp := int64(0)
	port := int64(443)

	if got := unsetInt64(&unsetIcmp, -1); !got.IsNull() {
		t.Errorf("icmp -1: got %v, want null", got)
	}
	if got := unsetInt64(&unsetPort, 0); !got.IsNull() {
		t.Errorf("port 0: got %v, want null", got)
	}
	if got := unsetInt64(nil, 0); !got.IsNull() {
		t.Errorf("nil: got %v, want null", got)
	}
	// icmp type 0 is a valid value and must be kept.
	if got := unsetInt64(&zeroIcmp, -1); got != types.Int64Value(0) {
		t.Errorf("icmp 0: got %v, want 0", got)
	}
	if got := unsetInt64(&port, 0); got != types.Int64Value(443) {
		t.Errorf("port 443: got %v, want 443", got)
	}

	empty := ""
	if got := unsetString(&empty); !got.IsNull() {
		t.Errorf("comment \"\": got %v, want null", got)
	}
	note := "Allow HTTP"
	if got := unsetString(&note); got != types.StringValue("Allow HTTP") {
		t.Errorf("comment: got %v, want Allow HTTP", got)
	}
}

func TestRuleFromAPIMapsProtocolDefaultToNull(t *testing.T) {
	t.Parallel()

	for _, p := range []interface{}{nil, "", "all"} {
		if got := ruleFromAPI(v9.FirewallRulesetRuleDetailResponse{Protocol: p}); !got.Protocol.IsNull() {
			t.Errorf("protocol %v: got %v, want null", p, got.Protocol)
		}
	}
	if got := ruleFromAPI(v9.FirewallRulesetRuleDetailResponse{Protocol: "tcp"}); got.Protocol != types.StringValue("tcp") {
		t.Errorf("protocol tcp: got %v, want tcp", got.Protocol)
	}
}

func TestKeepProtocolAll(t *testing.T) {
	t.Parallel()

	all := types.StringValue("all")
	rules := []ruleBlockModel{{Protocol: types.StringNull()}, {Protocol: types.StringNull()}, {Protocol: types.StringNull()}}
	prior := []ruleBlockModel{{Protocol: all}, {Protocol: types.StringNull()}}

	keepProtocolAll(rules, prior)

	if rules[0].Protocol != all {
		t.Errorf("rule 0: got %v, want all", rules[0].Protocol)
	}
	if !rules[1].Protocol.IsNull() {
		t.Errorf("rule 1: got %v, want null", rules[1].Protocol)
	}
	// A rule without a prior rule stays null.
	if !rules[2].Protocol.IsNull() {
		t.Errorf("rule 2: got %v, want null", rules[2].Protocol)
	}
}
