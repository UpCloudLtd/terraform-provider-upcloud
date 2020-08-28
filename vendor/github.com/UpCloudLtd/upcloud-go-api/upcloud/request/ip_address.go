package request

import (
	"encoding/json"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
)

// GetIPAddressDetailsRequest represents a request to retrieve details about a specific IP address
type GetIPAddressDetailsRequest struct {
	Address string
}

// RequestURL implements the Request interface
func (r *GetIPAddressDetailsRequest) RequestURL() string {
	return fmt.Sprintf("/ip_address/%s", r.Address)
}

// AssignIPAddressRequest represents a request to assign a new IP address to a server
type AssignIPAddressRequest struct {
	Access     string          `json:"access,omitempty"`
	Family     string          `json:"family,omitempty"`
	ServerUUID string          `json:"server,omitempty"`
	Floating   upcloud.Boolean `json:"floating,omitempty"`
	MAC        string          `json:"mac,omitempty"`
	Zone       string          `json:"zone,omitempty"`
}

// RequestURL implements the Request interface
func (r *AssignIPAddressRequest) RequestURL() string {
	return "/ip_address"
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r AssignIPAddressRequest) MarshalJSON() ([]byte, error) {
	type localAssignIPAddressRequest AssignIPAddressRequest
	v := struct {
		AssignIPAddressRequest localAssignIPAddressRequest `json:"ip_address"`
	}{}
	v.AssignIPAddressRequest = localAssignIPAddressRequest(r)

	return json.Marshal(&v)
}

// ModifyIPAddressRequest represents a request to modify the PTR DNS record of a specific IP address
type ModifyIPAddressRequest struct {
	IPAddress string `json:"-"`

	PTRRecord string `json:"ptr_record,omitempty"`
	MAC       string `json:"mac,omitempty"`
}

// RequestURL implements the Request interface
func (r *ModifyIPAddressRequest) RequestURL() string {
	return fmt.Sprintf("/ip_address/%s", r.IPAddress)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r ModifyIPAddressRequest) MarshalJSON() ([]byte, error) {
	if r.PTRRecord == "" && r.MAC == "" {
		// We want to unassign the IP which requires the MAC to be explicitly not set.
		return []byte(`
		  {
			"ip_address": {
			  "mac": null
			}
		  }
		`), nil
	}
	type localModifyIPAddressRequest ModifyIPAddressRequest
	v := struct {
		ModifyIPAddressRequest localModifyIPAddressRequest `json:"ip_address"`
	}{}
	v.ModifyIPAddressRequest = localModifyIPAddressRequest(r)

	return json.Marshal(&v)
}

// ReleaseIPAddressRequest represents a request to remove a specific IP address from server
type ReleaseIPAddressRequest struct {
	IPAddress string
}

// RequestURL implements the Request interface
func (r *ReleaseIPAddressRequest) RequestURL() string {
	return fmt.Sprintf("/ip_address/%s", r.IPAddress)
}
