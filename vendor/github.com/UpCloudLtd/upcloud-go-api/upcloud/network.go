package upcloud

import "encoding/json"

// Constants
const (
	NetworkTypePrivate = "private"
	NetworkTypePublic  = "public"
	NetworkTypeUtility = "utility"
)

// ServerInterface represent a network interface on the server
type ServerInterface Interface

// ServerInterfaceSlice is a slice of ServerInterfaces.
// It exists to allow for a custom JSON unmarshaller.
type ServerInterfaceSlice []ServerInterface

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *ServerInterfaceSlice) UnmarshalJSON(b []byte) error {
	v := struct {
		Interfaces []ServerInterface `json:"interface"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = v.Interfaces

	return nil
}

// Networking represents networking in a response
type Networking struct {
	Interfaces ServerInterfaceSlice `json:"interfaces"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *Networking) UnmarshalJSON(b []byte) error {
	type localNetworking Networking

	v := struct {
		Networking localNetworking `json:"networking"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = Networking(v.Networking)

	return nil
}

// Interface represents a network interface in a response
type Interface struct {
	Index             int            `json:"index"`
	IPAddresses       IPAddressSlice `json:"ip_addresses"`
	MAC               string         `json:"mac"`
	Network           string         `json:"network"`
	Type              string         `json:"type"`
	Bootable          Boolean        `json:"bootable"`
	SourceIPFiltering Boolean        `json:"source_ip_filtering"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *Interface) UnmarshalJSON(b []byte) error {
	type localInterface Interface

	v := struct {
		Interface localInterface `json:"interface"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = Interface(v.Interface)

	return nil
}

// IPNetwork represents an IP network in a response.
type IPNetwork struct {
	Address          string   `json:"address,omitempty"`
	DHCP             Boolean  `json:"dhcp"`
	DHCPDefaultRoute Boolean  `json:"dhcp_default_route"`
	DHCPDns          []string `json:"dhcp_dns,omitempty"`
	Family           string   `json:"family,omitempty"`
	Gateway          string   `json:"gateway,omitempty"`
}

// IPNetworkSlice is a slice of IPNetworks
// It exists to allow for a custom unmarshaller.
type IPNetworkSlice []IPNetwork

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (t *IPNetworkSlice) UnmarshalJSON(b []byte) error {
	v := struct {
		IPNetworks []IPNetwork `json:"ip_network"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*t) = v.IPNetworks

	return nil
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (t IPNetworkSlice) MarshalJSON() ([]byte, error) {
	v := struct {
		IPNetworks []IPNetwork `json:"ip_network"`
	}{}
	if t == nil {
		t = make(IPNetworkSlice, 0)
	}
	v.IPNetworks = t

	return json.Marshal(v)
}

// NetworkServerSlice is a slice of NetworkServers.
// It exists to allow for a custom JSON unmarshaller.
type NetworkServerSlice []NetworkServer

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (t *NetworkServerSlice) UnmarshalJSON(b []byte) error {
	v := struct {
		NetworkServers []NetworkServer `json:"server"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*t) = v.NetworkServers

	return nil
}

// NetworkServer represents a server in a networking response
type NetworkServer struct {
	ServerUUID  string `json:"uuid"`
	ServerTitle string `json:"title"`
}

// Network represents a network in a networking response.
type Network struct {
	IPNetworks IPNetworkSlice     `json:"ip_networks"`
	Name       string             `json:"name"`
	Type       string             `json:"type"`
	UUID       string             `json:"uuid"`
	Zone       string             `json:"zone"`
	Router     string             `json:"router"`
	Servers    NetworkServerSlice `json:"servers"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *Network) UnmarshalJSON(b []byte) error {
	type localNetwork Network

	v := struct {
		Network localNetwork `json:"network"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = Network(v.Network)

	return nil
}

// Networks represents multiple networks in a GetNetworks and GetNetworksInZone response.
type Networks struct {
	Networks []Network `json:"networks"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (n *Networks) UnmarshalJSON(b []byte) error {
	type localNetwork Network
	type networkWrapper struct {
		Networks []localNetwork `json:"network"`
	}

	v := struct {
		Networks networkWrapper `json:"networks"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	for _, ln := range v.Networks.Networks {
		n.Networks = append(n.Networks, Network(ln))
	}

	return nil
}

// NetworkSlice is a slice of Networks.
// It exists to allow for a custom JSON unmarshaller.
type NetworkSlice []Network

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (t *NetworkSlice) UnmarshalJSON(b []byte) error {
	v := struct {
		Networks []Network `json:"network"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*t) = v.Networks

	return nil
}

// Routers represents a response to a GetRouters request
type Routers struct {
	Routers []Router `json:"routers"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (n *Routers) UnmarshalJSON(b []byte) error {
	type localRouter Router
	type routerWrapper struct {
		Routers []localRouter `json:"router"`
	}

	v := struct {
		Routers routerWrapper `json:"routers"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	for _, ln := range v.Routers.Routers {
		n.Routers = append(n.Routers, Router(ln))
	}

	return nil
}

// RouterNetwork represents the networks in a router response.
type RouterNetwork struct {
	NetworkUUID string `json:"uuid"`
}

// RouterNetworkSlice is a slice of RouterNetworks.
// It exists to allow for a custom unmarshaller.
type RouterNetworkSlice []RouterNetwork

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (t *RouterNetworkSlice) UnmarshalJSON(b []byte) error {
	v := struct {
		Networks []RouterNetwork `json:"network"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*t) = v.Networks

	return nil
}

// Router represents a Router in a response
type Router struct {
	AttachedNetworks RouterNetworkSlice `json:"attached_networks"`
	Name             string             `json:"name"`
	Type             string             `json:"type"`
	UUID             string             `json:"uuid"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *Router) UnmarshalJSON(b []byte) error {
	type localRouter Router

	v := struct {
		Router localRouter `json:"router"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = Router(v.Router)

	return nil
}
