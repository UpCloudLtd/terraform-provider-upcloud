package upcloud

import "encoding/json"

// Constants
const (
	FirewallRuleActionAccept = "accept"
	FirewallRuleActionReject = "reject"
	FirewallRuleActionDrop   = "drop"

	FirewallRuleDirectionIn  = "in"
	FirewallRuleDirectionOut = "out"

	FirewallRuleProtocolTCP  = "tcp"
	FirewallRuleProtocolUDP  = "udp"
	FirewallRuleProtocolICMP = "icmp"
)

// FirewallRules represents a list of firewall rules
type FirewallRules struct {
	FirewallRules []FirewallRule `json:"firewall_rules"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *FirewallRules) UnmarshalJSON(b []byte) error {
	type localFirewallRule FirewallRule
	type firewallRuleWrapper struct {
		FirewallRules []localFirewallRule `json:"firewall_rule"`
	}

	v := struct {
		FirewallRules firewallRuleWrapper `json:"firewall_rules"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	for _, f := range v.FirewallRules.FirewallRules {
		s.FirewallRules = append(s.FirewallRules, FirewallRule(f))
	}

	return nil
}

// FirewallRule represents a single firewall rule. Note that most integer values are represented as strings
type FirewallRule struct {
	Action                  string `json:"action"`
	Comment                 string `json:"comment,omitempty"`
	DestinationAddressStart string `json:"destination_address_start,omitempty"`
	DestinationAddressEnd   string `json:"destination_address_end,omitempty"`
	DestinationPortStart    string `json:"destination_port_start,omitempty"`
	DestinationPortEnd      string `json:"destination_port_end,omitempty"`
	Direction               string `json:"direction"`
	Family                  string `json:"family"`
	ICMPType                string `json:"icmp_type,omitempty"`
	Position                int    `json:"position,string,omitempty"`
	Protocol                string `json:"protocol,omitempty"`
	SourceAddressStart      string `json:"source_address_start,omitempty"`
	SourceAddressEnd        string `json:"source_address_end,omitempty"`
	SourcePortStart         string `json:"source_port_start,omitempty"`
	SourcePortEnd           string `json:"source_port_end,omitempty"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *FirewallRule) UnmarshalJSON(b []byte) error {
	type localFirewallRule FirewallRule

	v := struct {
		FirewallRule localFirewallRule `json:"firewall_rule"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = FirewallRule(v.FirewallRule)

	return nil
}
