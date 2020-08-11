package request

import (
	"encoding/json"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
)

// GetFirewallRulesRequest represents a request for retrieving the firewall rules for a specific server
type GetFirewallRulesRequest struct {
	ServerUUID string
}

// RequestURL implements the Request interface
func (r *GetFirewallRulesRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/firewall_rule", r.ServerUUID)
}

// GetFirewallRuleDetailsRequest represents a request to get details about a specific firewall rule
type GetFirewallRuleDetailsRequest struct {
	ServerUUID string
	Position   int
}

// RequestURL implements the Request interface
func (r *GetFirewallRuleDetailsRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/firewall_rule/%d", r.ServerUUID, r.Position)
}

// CreateFirewallRuleRequest represents a request to create a new firewall rule for a specific server
type CreateFirewallRuleRequest struct {
	upcloud.FirewallRule

	ServerUUID string `json:"-"`
}

// RequestURL implements the Request interface
func (r *CreateFirewallRuleRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/firewall_rule", r.ServerUUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r CreateFirewallRuleRequest) MarshalJSON() ([]byte, error) {
	type localCreateFirewallRuleRequest CreateFirewallRuleRequest
	v := struct {
		CreateFirewallRuleRequest localCreateFirewallRuleRequest `json:"firewall_rule"`
	}{}
	v.CreateFirewallRuleRequest = localCreateFirewallRuleRequest(r)

	return json.Marshal(&v)
}

// DeleteFirewallRuleRequest represents a request to remove a firewall rule
type DeleteFirewallRuleRequest struct {
	ServerUUID string
	Position   int
}

// RequestURL implements the Request interface
func (r *DeleteFirewallRuleRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/firewall_rule/%d", r.ServerUUID, r.Position)
}
