package upcloud

import "encoding/json"

// Constants
const (
	IPAddressFamilyIPv4 = "IPv4"
	IPAddressFamilyIPv6 = "IPv6"

	IPAddressAccessPrivate = "private"
	IPAddressAccessPublic  = "public"
)

// IPAddresses represents a /ip_address response
type IPAddresses struct {
	IPAddresses []IPAddress `json:"ip_addresses"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *IPAddresses) UnmarshalJSON(b []byte) error {
	type localIPAddress IPAddress
	type ipAddressWrapper struct {
		IPAddresses []localIPAddress `json:"ip_address"`
	}

	v := struct {
		IPAddresses ipAddressWrapper `json:"ip_addresses"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	for _, ip := range v.IPAddresses.IPAddresses {
		s.IPAddresses = append(s.IPAddresses, IPAddress(ip))
	}

	return nil
}

// IPAddress represents an IP address
type IPAddress struct {
	Access  string `json:"access"`
	Address string `json:"address"`
	Family  string `json:"family"`
	// TODO: Convert to boolean
	PartOfPlan string `json:"part_of_plan"`
	PTRRecord  string `json:"ptr_record"`
	ServerUUID string `json:"server"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *IPAddress) UnmarshalJSON(b []byte) error {
	type localIPAddress IPAddress

	v := struct {
		IPAddress localIPAddress `json:"ip_address"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = IPAddress(v.IPAddress)

	return nil
}
