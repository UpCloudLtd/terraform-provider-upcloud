package request

import (
	"encoding/xml"
	"fmt"
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
	XMLName xml.Name `xml:"ip_address"`

	Access     string `xml:"access"`
	Family     string `xml:"family,omitempty"`
	ServerUUID string `xml:"server"`
}

// RequestURL implements the Request interface
func (r *AssignIPAddressRequest) RequestURL() string {
	return "/ip_address"
}

// ModifyIPAddressRequest represents a request to modify the PTR DNS record of a specific IP address
type ModifyIPAddressRequest struct {
	XMLName   xml.Name `xml:"ip_address"`
	IPAddress string   `xml:"-"`

	PTRRecord string `xml:"ptr_record"`
}

// RequestURL implements the Request interface
func (r *ModifyIPAddressRequest) RequestURL() string {
	return fmt.Sprintf("/ip_address/%s", r.IPAddress)
}

// ReleaseIPAddressRequest represents a request to remove a specific IP address from server
type ReleaseIPAddressRequest struct {
	IPAddress string
}

// RequestURL implements the Request interface
func (r *ReleaseIPAddressRequest) RequestURL() string {
	return fmt.Sprintf("/ip_address/%s", r.IPAddress)
}
