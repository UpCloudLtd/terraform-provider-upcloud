package upcloud

import (
	"encoding/json"
)

// Constants
const (
	ServerStateStarted     = "started"
	ServerStateStopped     = "stopped"
	ServerStateMaintenance = "maintenance"
	ServerStateError       = "error"

	VideoModelVGA    = "vga"
	VideoModelCirrus = "cirrus"

	StopTypeSoft = "soft"
	StopTypeHard = "hard"

	RemoteAccessTypeVNC   = "vnc"
	RemoteAccessTypeSPICE = "spice"
)

// ServerConfigurations represents a /server_size response
type ServerConfigurations struct {
	ServerConfigurations []ServerConfiguration `json:"server_sizes"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *ServerConfigurations) UnmarshalJSON(b []byte) error {
	type serverConfigurationWrapper struct {
		ServerConfigurations []ServerConfiguration `json:"server_size"`
	}

	v := struct {
		ServerConfigurations serverConfigurationWrapper `json:"server_sizes"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	s.ServerConfigurations = v.ServerConfigurations.ServerConfigurations

	return nil
}

// ServerConfiguration represents a server configuration
type ServerConfiguration struct {
	CoreNumber   int `json:"core_number,string"`
	MemoryAmount int `json:"memory_amount,string"`
}

// Servers represents a /server response
type Servers struct {
	Servers []Server `json:"servers"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *Servers) UnmarshalJSON(b []byte) error {
	type serverWrapper struct {
		Servers []Server `json:"server"`
	}

	v := struct {
		Servers serverWrapper `json:"servers"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	s.Servers = v.Servers.Servers

	return nil
}

// ServerTagSlice is a slice of string.
// It exists to allow for a custom JSON unmarshaller.
type ServerTagSlice []string

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (t *ServerTagSlice) UnmarshalJSON(b []byte) error {
	v := struct {
		Tags []string `json:"tag"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*t) = v.Tags

	return nil
}

// Server represents a server
type Server struct {
	CoreNumber   int            `json:"core_number,string"`
	Hostname     string         `json:"hostname"`
	License      float64        `json:"license"`
	MemoryAmount int            `json:"memory_amount,string"`
	Plan         string         `json:"plan"`
	Progress     int            `json:"progress,string"`
	State        string         `json:"state"`
	Tags         ServerTagSlice `json:"tags"`
	Title        string         `json:"title"`
	UUID         string         `json:"uuid"`
	Zone         string         `json:"zone"`
}

// IPAddressSlice is a slice of IPAddress.
// It exists to allow for a custom JSON unmarshaller.
type IPAddressSlice []IPAddress

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (i *IPAddressSlice) UnmarshalJSON(b []byte) error {
	type localIPAddress IPAddress
	v := struct {
		IPAddresses []localIPAddress `json:"ip_address"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	for _, ip := range v.IPAddresses {
		(*i) = append((*i), IPAddress(ip))
	}

	return nil
}

// ServerStorageDeviceSlice is a slice of ServerStorageDevices.
// It exists to allow for a custom JSON unmarshaller.
type ServerStorageDeviceSlice []ServerStorageDevice

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *ServerStorageDeviceSlice) UnmarshalJSON(b []byte) error {
	v := struct {
		StorageDevices []ServerStorageDevice `json:"storage_device"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = v.StorageDevices

	return nil
}

// ServerNetworking represents the networking on a server response.
// It is castable to a Networking struct.
type ServerNetworking Networking

// ServerDetails represents details about a server
type ServerDetails struct {
	Server

	BootOrder string `json:"boot_order"`
	// TODO: Convert to boolean
	Firewall             string                   `json:"firewall"`
	Host                 int                      `json:"host"`
	IPAddresses          IPAddressSlice           `json:"ip_addresses"`
	Metadata             Boolean                  `json:"metadata"`
	NICModel             string                   `json:"nic_model"`
	Networking           ServerNetworking         `json:"networking"`
	SimpleBackup         string                   `json:"simple_backup"`
	StorageDevices       ServerStorageDeviceSlice `json:"storage_devices"`
	Timezone             string                   `json:"timezone"`
	VideoModel           string                   `json:"video_model"`
	RemoteAccessEnabled  Boolean                  `json:"remote_access_enabled"`
	RemoteAccessType     string                   `json:"remote_access_type"`
	RemoteAccessHost     string                   `json:"remote_access_host"`
	RemoteAccessPassword string                   `json:"remote_access_password"`
	RemoteAccessPort     int                      `json:"remote_access_port,string"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *ServerDetails) UnmarshalJSON(b []byte) error {
	type localServerDetails ServerDetails

	v := struct {
		ServerDetails localServerDetails `json:"server"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = ServerDetails(v.ServerDetails)

	return nil
}
