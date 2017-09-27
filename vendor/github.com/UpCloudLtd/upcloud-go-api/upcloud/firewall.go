package upcloud

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
	FirewallRules []FirewallRule `xml:"firewall_rule"`
}

// FirewallRule represents a single firewall rule. Note that most integer values are represented as strings
type FirewallRule struct {
	Action                  string `xml:"action"`
	Comment                 string `xml:"comment,omitempty"`
	DestinationAddressStart string `xml:"destination_address_start,omitempty"`
	DestinationAddressEnd   string `xml:"destination_address_end,omitempty"`
	DestinationPortStart    string `xml:"destination_port_start,omitempty"`
	DestinationPortEnd      string `xml:"destination_port_end,omitempty"`
	Direction               string `xml:"direction"`
	Family                  string `xml:"family"`
	ICMPType                string `xml:"icmp_type,omitempty"`
	Position                int    `xml:"position"`
	Protocol                string `xml:"protocol,omitempty"`
	SourceAddressStart      string `xml:"source_address_start,omitempty"`
	SourceAddressEnd        string `xml:"source_address_end,omitempty"`
	SourcePortStart         string `xml:"source_port_start,omitempty"`
	SourcePortEnd           string `xml:"source_port_end,omitempty"`
}
