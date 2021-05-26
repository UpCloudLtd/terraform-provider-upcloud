package request

import (
	"encoding/json"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
)

// GetNetworksInZoneRequest represents a request to get all networks
// within the specified zone.
type GetNetworksInZoneRequest struct {
	Zone string
}

// RequestURL implements the Request interface.
func (r *GetNetworksInZoneRequest) RequestURL() string {
	return fmt.Sprintf("/network/?zone=%s", r.Zone)
}

// GetNetworkDetailsRequest represents a request to the the details of
// a single network.
type GetNetworkDetailsRequest struct {
	UUID string
}

// RequestURL implements the Request interface.
func (r *GetNetworkDetailsRequest) RequestURL() string {
	return fmt.Sprintf("/network/%s", r.UUID)
}

// CreateNetworkRequest represents a request to create a new network.
type CreateNetworkRequest struct {
	Name       string                 `json:"name,omitempty"`
	Zone       string                 `json:"zone,omitempty"`
	Router     string                 `json:"router,omitempty"`
	IPNetworks upcloud.IPNetworkSlice `json:"ip_networks,omitempty"`
}

// RequestURL implements the Request interface.
func (r *CreateNetworkRequest) RequestURL() string {
	return "/network/"
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r CreateNetworkRequest) MarshalJSON() ([]byte, error) {
	type localCreateNetworkRequest CreateNetworkRequest
	v := struct {
		CreateNetworkRequest localCreateNetworkRequest `json:"network"`
	}{}
	v.CreateNetworkRequest = localCreateNetworkRequest(r)

	return json.Marshal(&v)
}

// ModifyNetworkRequest represents a request to modify an existing network.
type ModifyNetworkRequest struct {
	UUID string `json:"-"`

	Name       string                 `json:"name,omitempty"`
	Zone       string                 `json:"zone,omitempty"`
	IPNetworks upcloud.IPNetworkSlice `json:"ip_networks,omitempty"`
}

// RequestURL implements the Request interface.
func (r *ModifyNetworkRequest) RequestURL() string {
	return fmt.Sprintf("/network/%s", r.UUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r ModifyNetworkRequest) MarshalJSON() ([]byte, error) {
	type localModifyNetworkRequest ModifyNetworkRequest
	v := struct {
		ModifyNetworkRequest localModifyNetworkRequest `json:"network"`
	}{}
	v.ModifyNetworkRequest = localModifyNetworkRequest(r)

	return json.Marshal(&v)
}

// DeleteNetworkRequest represents a request to delete a network.
type DeleteNetworkRequest struct {
	UUID string
}

// RequestURL implements the Request interface.
func (r *DeleteNetworkRequest) RequestURL() string {
	return fmt.Sprintf("/network/%s", r.UUID)
}

// AttachNetworkRouterRequest represents a request to attach a particular router to a network
type AttachNetworkRouterRequest struct {
	NetworkUUID string `json:"-"`
	RouterUUID  string `json:"router"`
}

// RequestURL implements the Request interface
func (r *AttachNetworkRouterRequest) RequestURL() string {
	return (&ModifyNetworkRequest{UUID: r.NetworkUUID}).RequestURL()
}

// MarshalJSON implements the json.Marshaler interface
func (r AttachNetworkRouterRequest) MarshalJSON() ([]byte, error) {
	type localAttachNetworkRouterRequest AttachNetworkRouterRequest
	v := struct {
		AttachNetworkRouterRequest localAttachNetworkRouterRequest `json:"network"`
	}{}
	v.AttachNetworkRouterRequest = localAttachNetworkRouterRequest(r)

	return json.Marshal(&v)
}

// DetachNetworkRouterRequest represents a request to detach a router from a network
type DetachNetworkRouterRequest struct {
	NetworkUUID string `json:"-"`
}

// RequestURL implements the Request interface
func (r *DetachNetworkRouterRequest) RequestURL() string {
	return (&ModifyNetworkRequest{UUID: r.NetworkUUID}).RequestURL()
}

// MarshalJSON implements the json.Marshaler interface
func (r DetachNetworkRouterRequest) MarshalJSON() ([]byte, error) {
	return []byte(`{ "network": { "router": null } }`), nil
}

// GetServerNetworksRequest represents a request to get the networks
// a server is part of.
type GetServerNetworksRequest struct {
	ServerUUID string
}

// RequestURL implements the Request interface.
func (r *GetServerNetworksRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/networking", r.ServerUUID)
}

// CreateNetworkInterfaceIPAddress represents an IP Address object
// that is needed to create a network interface.
type CreateNetworkInterfaceIPAddress struct {
	Family  string `json:"family"`
	Address string `json:"address,omitempty"`
}

// CreateNetworkInterfaceIPAddressSlice is a slice of
// CreateNetworkInterfaceIPAddress.
// It exists to allow for a custom JSON marshaller.
type CreateNetworkInterfaceIPAddressSlice []CreateNetworkInterfaceIPAddress

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (s CreateNetworkInterfaceIPAddressSlice) MarshalJSON() ([]byte, error) {
	v := struct {
		IPAddress []CreateNetworkInterfaceIPAddress `json:"ip_address"`
	}{}
	v.IPAddress = s

	return json.Marshal(v)
}

// CreateNetworkInterfaceRequest represents a request to create a new network
// interface on a server.
type CreateNetworkInterfaceRequest struct {
	ServerUUID string `json:"-"`

	Type              string                               `json:"type"`
	NetworkUUID       string                               `json:"network,omitempty"`
	Index             int                                  `json:"index,omitempty"`
	IPAddresses       CreateNetworkInterfaceIPAddressSlice `json:"ip_addresses"`
	SourceIPFiltering upcloud.Boolean                      `json:"source_ip_filtering,omitempty"`
	Bootable          upcloud.Boolean                      `json:"bootable,omitempty"`
}

// RequestURL implements the Request interface.
func (r *CreateNetworkInterfaceRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/networking/interface", r.ServerUUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r CreateNetworkInterfaceRequest) MarshalJSON() ([]byte, error) {
	type localCreateNetworkInterfaceRequest CreateNetworkInterfaceRequest
	v := struct {
		CreateNetworkInterfaceRequest localCreateNetworkInterfaceRequest `json:"interface"`
	}{}
	v.CreateNetworkInterfaceRequest = localCreateNetworkInterfaceRequest(r)

	return json.Marshal(&v)
}

// DeleteNetworkInterfaceRequest represents a request to delete a network interface from a server.
type DeleteNetworkInterfaceRequest struct {
	ServerUUID string
	Index      int
}

// RequestURL implements the Request interface.
func (r *DeleteNetworkInterfaceRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/networking/interface/%d", r.ServerUUID, r.Index)
}

// ModifyNetworkInterfaceRequest represents a request to modify a network interface on a server.
type ModifyNetworkInterfaceRequest struct {
	ServerUUID   string `json:"-"`
	CurrentIndex int    `json:"-"`

	Type              string                               `json:"type,omitempty"`
	NetworkUUID       string                               `json:"network,omitempty"`
	NewIndex          int                                  `json:"index,omitempty"`
	IPAddresses       CreateNetworkInterfaceIPAddressSlice `json:"ip_addresses,omitempty"`
	SourceIPFiltering upcloud.Boolean                      `json:"source_ip_filtering,omitempty"`
	Bootable          upcloud.Boolean                      `json:"bootable,omitempty"`
}

// RequestURL implements the Request interface.
func (r *ModifyNetworkInterfaceRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/networking/interface/%d", r.ServerUUID, r.CurrentIndex)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r ModifyNetworkInterfaceRequest) MarshalJSON() ([]byte, error) {
	type localModifyNetworkInterfaceRequest ModifyNetworkInterfaceRequest
	v := struct {
		ModifyNetworkInterfaceRequest localModifyNetworkInterfaceRequest `json:"interface"`
	}{}
	v.ModifyNetworkInterfaceRequest = localModifyNetworkInterfaceRequest(r)

	return json.Marshal(&v)
}

// GetRouterDetailsRequest represents a request to get details about a single router.
type GetRouterDetailsRequest struct {
	UUID string
}

// RequestURL implements the Request interface.
func (r *GetRouterDetailsRequest) RequestURL() string {
	return fmt.Sprintf("/router/%s", r.UUID)
}

// CreateRouterRequest represents a request to create a new router.
type CreateRouterRequest struct {
	Name string `json:"name"`
}

// RequestURL implements the Request interface.
func (r *CreateRouterRequest) RequestURL() string {
	return "/router"
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r CreateRouterRequest) MarshalJSON() ([]byte, error) {
	type localCreateRouterRequest CreateRouterRequest
	v := struct {
		CreateRouterRequest localCreateRouterRequest `json:"router"`
	}{}
	v.CreateRouterRequest = localCreateRouterRequest(r)

	return json.Marshal(&v)
}

// ModifyRouterRequest represents a request to modify an existing router.
type ModifyRouterRequest struct {
	UUID string `json:"-"`

	Name string `json:"name"`
}

// RequestURL implements the Request interface.
func (r *ModifyRouterRequest) RequestURL() string {
	return fmt.Sprintf("/router/%s", r.UUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r ModifyRouterRequest) MarshalJSON() ([]byte, error) {
	type localModifyRouterRequest ModifyRouterRequest
	v := struct {
		ModifyRouterRequest localModifyRouterRequest `json:"router"`
	}{}
	v.ModifyRouterRequest = localModifyRouterRequest(r)

	return json.Marshal(&v)
}

// DeleteRouterRequest represents a request to delete a router.
type DeleteRouterRequest struct {
	UUID string
}

// RequestURL implements the Request interface.
func (r *DeleteRouterRequest) RequestURL() string {
	return fmt.Sprintf("/router/%s", r.UUID)
}
