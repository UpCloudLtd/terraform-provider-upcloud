package request

import (
	"encoding/xml"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestGetFirewallRulesRequest tests that GetFirewallRulesRequest behaves correctly
func TestGetFirewallRulesRequest(t *testing.T) {
	request := GetFirewallRulesRequest{
		ServerUUID: "00798b85-efdc-41ca-8021-f6ef457b8531",
	}

	assert.Equal(t, "/server/00798b85-efdc-41ca-8021-f6ef457b8531/firewall_rule", request.RequestURL())
}

// TestGetFirewallRuleDetailsRequest tests that GetFirewallRuleDetailsRequest behaves correctly
func TestGetFirewallRuleDetailsRequest(t *testing.T) {
	request := GetFirewallRuleDetailsRequest{
		ServerUUID: "00798b85-efdc-41ca-8021-f6ef457b8531",
		Position:   1,
	}

	assert.Equal(t, "/server/00798b85-efdc-41ca-8021-f6ef457b8531/firewall_rule/1", request.RequestURL())
}

// TestCreateFirewallRuleRequest tests that CreateFirewallRuleRequest behaves correctly
func TestCreateFirewallRuleRequest(t *testing.T) {
	request := CreateFirewallRuleRequest{
		ServerUUID: "00798b85-efdc-41ca-8021-f6ef457b8531",
		FirewallRule: upcloud.FirewallRule{
			Direction: upcloud.FirewallRuleDirectionIn,
			Action:    upcloud.FirewallRuleActionAccept,
			Family:    upcloud.IPAddressFamilyIPv4,
			Position:  1,
			Comment:   "This is the comment",
		},
	}

	// Check the request URL
	assert.Equal(t, "/server/00798b85-efdc-41ca-8021-f6ef457b8531/firewall_rule", request.RequestURL())

	// Check marshaling
	byteXML, err := xml.Marshal(&request)
	assert.Nil(t, err)

	expectedXML := "<firewall_rule><action>accept</action><comment>This is the comment</comment><direction>in</direction><family>IPv4</family><position>1</position></firewall_rule>"
	actualXML := string(byteXML)
	assert.Equal(t, expectedXML, actualXML)
}

// TestDeleteFirewallRuleRequest tests that DeleteFirewallRuleRequest behaves correctly
func TestDeleteFirewallRuleRequest(t *testing.T) {
	request := DeleteFirewallRuleRequest{
		ServerUUID: "00798b85-efdc-41ca-8021-f6ef457b8531",
		Position:   1,
	}

	assert.Equal(t, "/server/00798b85-efdc-41ca-8021-f6ef457b8531/firewall_rule/1", request.RequestURL())
}
