package firewall

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestFirewallRuleValidateOptionalPort(t *testing.T) {
	p := cty.Path{}
	if diag := firewallRuleValidateOptionalPort("1", p); len(diag) > 0 {
		t.Error(diag[0].Detail)
	}

	if diag := firewallRuleValidateOptionalPort("65535", p); len(diag) > 0 {
		t.Error(diag[0].Detail)
	}

	if diag := firewallRuleValidateOptionalPort("abc", p); len(diag) < 1 {
		t.Error("firewallRuleValidateOptionalPort failed 'abc' is not valid port")
	}

	if diag := firewallRuleValidateOptionalPort("0", p); len(diag) < 1 {
		t.Error("firewallRuleValidateOptionalPort failed '0' is not valid port")
	}

	if diag := firewallRuleValidateOptionalPort("65536", p); len(diag) < 1 {
		t.Error("firewallRuleValidateOptionalPort failed '65536' is not valid port")
	}
}
