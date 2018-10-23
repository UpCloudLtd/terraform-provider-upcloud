package upcloud

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUnmarshalFirewallRules tests the FirewallRules and FirewallRule are unmarshaled correctly
func TestUnmarshalFirewallRules(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="utf-8"?>
<firewall_rules>
    <firewall_rule>
        <action>accept</action>
        <comment>HTTP 80</comment>
        <destination_address_end></destination_address_end>
        <destination_address_start></destination_address_start>
        <destination_port_end>80</destination_port_end>
        <destination_port_start>80</destination_port_start>
        <direction>in</direction>
        <family>IPv4</family>
        <icmp_type></icmp_type>
        <position>1</position>
        <protocol>tcp</protocol>
        <source_address_end></source_address_end>
        <source_address_start></source_address_start>
        <source_port_end></source_port_end>
        <source_port_start></source_port_start>
    </firewall_rule>
    <firewall_rule>
        <action>reject</action>
        <comment>ICMP</comment>
        <destination_address_end></destination_address_end>
        <destination_address_start></destination_address_start>
        <destination_port_end></destination_port_end>
        <destination_port_start></destination_port_start>
        <direction>in</direction>
        <family>IPv6</family>
        <icmp_type>1</icmp_type>
        <position>2</position>
        <protocol>icmp</protocol>
        <source_address_end></source_address_end>
        <source_address_start></source_address_start>
        <source_port_end></source_port_end>
        <source_port_start></source_port_start>
    </firewall_rule>
</firewall_rules>`

	firewallRules := FirewallRules{}
	err := xml.Unmarshal([]byte(originalXML), &firewallRules)
	assert.Nil(t, err)
	assert.Len(t, firewallRules.FirewallRules, 2)

	firstRule := firewallRules.FirewallRules[0]
	assert.Equal(t, FirewallRuleActionAccept, firstRule.Action)
	assert.Equal(t, "HTTP 80", firstRule.Comment)
	assert.Empty(t, firstRule.DestinationAddressEnd)
	assert.Empty(t, firstRule.DestinationAddressStart)
	assert.Equal(t, "80", firstRule.DestinationPortEnd)
	assert.Equal(t, "80", firstRule.DestinationPortStart)
	assert.Equal(t, FirewallRuleDirectionIn, firstRule.Direction)
	assert.Equal(t, IPAddressFamilyIPv4, firstRule.Family)
	assert.Empty(t, firstRule.ICMPType)
	assert.Equal(t, 1, firstRule.Position)
	assert.Equal(t, FirewallRuleProtocolTCP, firstRule.Protocol)
	assert.Empty(t, firstRule.SourceAddressEnd)
	assert.Empty(t, firstRule.SourceAddressStart)
	assert.Empty(t, firstRule.SourcePortEnd)
	assert.Empty(t, firstRule.SourcePortStart)
}
