package request

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
)

// Constants
const (
	PasswordDeliveryNone  = "none"
	PasswordDeliveryEmail = "email"
	PasswordDeliverySMS   = "sms"

	ServerStopTypeSoft = "soft"
	ServerStopTypeHard = "hard"

	RestartTimeoutActionDestroy = "destroy"
	RestartTimeoutActionIgnore  = "ignore"

	CreateServerStorageDeviceActionCreate = "create"
	CreateServerStorageDeviceActionClone  = "clone"
	CreateServerStorageDeviceActionAttach = "attach"
)

// GetServerDetailsRequest represents a request for retrieving details about a server
type GetServerDetailsRequest struct {
	UUID string
}

// RequestURL implements the Request interface
func (r *GetServerDetailsRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s", r.UUID)
}

// CreateServerIPAddressSlice is a slice of strings
// It exists to allow for a custom JSON marshaller.
type CreateServerIPAddressSlice []CreateServerIPAddress

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (s CreateServerIPAddressSlice) MarshalJSON() ([]byte, error) {
	v := struct {
		IPAddress []CreateServerIPAddress `json:"ip_address"`
	}{}
	v.IPAddress = s

	return json.Marshal(v)
}

// CreateServerStorageDevice represents a storage device for a CreateServerRequest
type CreateServerStorageDevice struct {
	Action  string `json:"action"`
	Address string `json:"address,omitempty"`
	Storage string `json:"storage"`
	Title   string `json:"title,omitempty"`
	// Storage size in gigabytes
	Size       int                 `json:"size,omitempty"`
	Tier       string              `json:"tier,omitempty"`
	Type       string              `json:"type,omitempty"`
	BackupRule *upcloud.BackupRule `json:"backup_rule,omitempty"`
}

// CreateServerStorageDeviceSlice is a slice of CreateServerStorageDevices
// It exists to allow for a custom JSON marshaller.
type CreateServerStorageDeviceSlice []CreateServerStorageDevice

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (s CreateServerStorageDeviceSlice) MarshalJSON() ([]byte, error) {
	v := struct {
		StorageDevice []CreateServerStorageDevice `json:"storage_device"`
	}{}
	v.StorageDevice = s

	return json.Marshal(v)
}

// CreateServerInterface represents a server network interface
// that is needed during server creation.
type CreateServerInterface struct {
	IPAddresses       CreateServerIPAddressSlice `json:"ip_addresses"`
	Type              string                     `json:"type"`
	Network           string                     `json:"network,omitempty"`
	SourceIPFiltering upcloud.Boolean            `json:"source_ip_filtering,omitempty"`
	Bootable          upcloud.Boolean            `json:"bootable,omitempty"`
}

// CreateServerInterfaceSlice is a slice of CreateServerInterfaces.
// It exists to allow for a custom JSON marshaller.
type CreateServerInterfaceSlice []CreateServerInterface

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (s CreateServerInterfaceSlice) MarshalJSON() ([]byte, error) {
	v := struct {
		Interfaces []CreateServerInterface `json:"interface"`
	}{}
	v.Interfaces = s

	return json.Marshal(v)
}

// CreateServerNetworking represents the networking details of a server
// needed during server creation.
type CreateServerNetworking struct {
	Interfaces CreateServerInterfaceSlice `json:"interfaces"`
}

// CreateServerRequest represents a request for creating a new server
type CreateServerRequest struct {
	AvoidHost  int    `json:"avoid_host,omitempty"`
	Host       int    `json:"host,omitempty"`
	BootOrder  string `json:"boot_order,omitempty"`
	CoreNumber int    `json:"core_number,omitempty"`
	// TODO: Convert to boolean
	Firewall             string                         `json:"firewall,omitempty"`
	Hostname             string                         `json:"hostname"`
	LoginUser            *LoginUser                     `json:"login_user,omitempty"`
	MemoryAmount         int                            `json:"memory_amount,omitempty"`
	Metadata             upcloud.Boolean                `json:"metadata"`
	Networking           *CreateServerNetworking        `json:"networking"`
	PasswordDelivery     string                         `json:"password_delivery,omitempty"`
	Plan                 string                         `json:"plan,omitempty"`
	SimpleBackup         string                         `json:"simple_backup,omitempty"`
	StorageDevices       CreateServerStorageDeviceSlice `json:"storage_devices"`
	TimeZone             string                         `json:"timezone,omitempty"`
	Title                string                         `json:"title"`
	UserData             string                         `json:"user_data,omitempty"`
	VideoModel           string                         `json:"video_model,omitempty"`
	RemoteAccessEnabled  upcloud.Boolean                `json:"remote_access_enabled"`
	RemoteAccessType     string                         `json:"remote_access_type,omitempty"`
	RemoteAccessPassword string                         `json:"remote_access_password,omitempty"`
	Zone                 string                         `json:"zone"`
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r CreateServerRequest) MarshalJSON() ([]byte, error) {
	type localCreateServerRequest CreateServerRequest
	v := struct {
		Server localCreateServerRequest `json:"server"`
	}{}
	v.Server = localCreateServerRequest(r)

	return json.Marshal(&v)
}

// RequestURL implements the Request interface
func (r *CreateServerRequest) RequestURL() string {
	return "/server"
}

// SSHKeySlice is a slice of strings
// It exists to allow for a custom JSON unmarshaller.
type SSHKeySlice []string

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *SSHKeySlice) UnmarshalJSON(b []byte) error {
	v := struct {
		SSHKey []string `json:"ssh_key"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = v.SSHKey

	return nil
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (s SSHKeySlice) MarshalJSON() ([]byte, error) {
	v := struct {
		SSHKey []string `json:"ssh_key"`
	}{}

	v.SSHKey = s

	return json.Marshal(v)
}

// LoginUser represents the login_user block when creating a new server
type LoginUser struct {
	CreatePassword string      `json:"create_password,omitempty"`
	Username       string      `json:"username,omitempty"`
	SSHKeys        SSHKeySlice `json:"ssh_keys,omitempty"`
}

// CreateServerIPAddress represents an IP address for a CreateServerRequest
type CreateServerIPAddress struct {
	Family string `json:"family"`
}

// WaitForServerStateRequest represents a request to wait for a server to enter or exit a specific state
type WaitForServerStateRequest struct {
	UUID           string
	DesiredState   string
	UndesiredState string
	Timeout        time.Duration
}

// StartServerRequest represents a request to start a server
type StartServerRequest struct {
	UUID string `json:"-"`

	// TODO: Start server requests have no timeout in the API
	Timeout time.Duration `json:"-"`

	AvoidHost int `json:"avoid_host,omitempty"`
	Host      int `json:"host,omitempty"`
}

// RequestURL implements the Request interface
func (r *StartServerRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/start", r.UUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r StartServerRequest) MarshalJSON() ([]byte, error) {
	type localStartServerRequest StartServerRequest
	v := struct {
		Server localStartServerRequest `json:"server"`
	}{}
	v.Server = localStartServerRequest(r)

	return json.Marshal(&v)
}

// StopServerRequest represents a request to stop a server
type StopServerRequest struct {
	UUID string `json:"-"`

	StopType string        `json:"stop_type,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty,string"`
}

// RequestURL implements the Request interface
func (r *StopServerRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/stop", r.UUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r StopServerRequest) MarshalJSON() ([]byte, error) {
	type localStopServerRequest StopServerRequest
	v := struct {
		StopServerRequest localStopServerRequest `json:"stop_server"`
	}{}
	v.StopServerRequest = localStopServerRequest(r)
	v.StopServerRequest.Timeout = v.StopServerRequest.Timeout / 1e9

	return json.Marshal(&v)
}

// RestartServerRequest represents a request to restart a server
type RestartServerRequest struct {
	UUID string `json:"-"`

	StopType      string        `json:"stop_type,omitempty"`
	Timeout       time.Duration `json:"timeout,omitempty,string"`
	TimeoutAction string        `json:"timeout_action,omitempty"`
	Host          int           `json:"host,omitempty"`
}

// RequestURL implements the Request interface
func (r *RestartServerRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/restart", r.UUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r RestartServerRequest) MarshalJSON() ([]byte, error) {
	type localRestartServerRequest RestartServerRequest
	v := struct {
		RestartServerRequest localRestartServerRequest `json:"restart_server"`
	}{}
	v.RestartServerRequest = localRestartServerRequest(r)
	v.RestartServerRequest.Timeout = v.RestartServerRequest.Timeout / 1e9

	return json.Marshal(&v)
}

// ModifyServerRequest represents a request to modify a server
type ModifyServerRequest struct {
	UUID string `json:"-"`

	BootOrder  string `json:"boot_order,omitempty"`
	CoreNumber int    `json:"core_number,omitempty,string"`
	// TODO: Convert to boolean
	Firewall             string          `json:"firewall,omitempty"`
	Hostname             string          `json:"hostname,omitempty"`
	MemoryAmount         int             `json:"memory_amount,omitempty,string"`
	Metadata             upcloud.Boolean `json:"metadata"`
	Plan                 string          `json:"plan,omitempty"`
	SimpleBackup         string          `json:"simple_backup,omitempty"`
	TimeZone             string          `json:"timezone,omitempty"`
	Title                string          `json:"title,omitempty"`
	VideoModel           string          `json:"video_model,omitempty"`
	RemoteAccessEnabled  upcloud.Boolean `json:"remote_access_enabled"`
	RemoteAccessType     string          `json:"remote_access_type,omitempty"`
	RemoteAccessPassword string          `json:"remote_access_password,omitempty"`
	Zone                 string          `json:"zone,omitempty"`
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r ModifyServerRequest) MarshalJSON() ([]byte, error) {
	type localModifyServerRequest ModifyServerRequest
	v := struct {
		ModifyServerRequest localModifyServerRequest `json:"server"`
	}{}
	v.ModifyServerRequest = localModifyServerRequest(r)

	return json.Marshal(&v)
}

// RequestURL implements the Request interface
func (r *ModifyServerRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s", r.UUID)
}

// DeleteServerRequest represents a request to delete a server
type DeleteServerRequest struct {
	UUID string
}

// RequestURL implements the Request interface
func (r *DeleteServerRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s", r.UUID)
}

// DeleteServerAndStoragesRequest represents a request to delete a server and all attached storages
type DeleteServerAndStoragesRequest struct {
	UUID string
}

// RequestURL implements the Request interface
func (r *DeleteServerAndStoragesRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/?storages=1", r.UUID)
}

// TagServerRequest represents a request to tag a server with one or more tags
type TagServerRequest struct {
	UUID string
	Tags []string
}

// RequestURL implements the Request interface
func (r *TagServerRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/tag/%s", r.UUID, strings.Join(r.Tags, ","))
}

// UntagServerRequest represents a request to remove one or more tags from a server
type UntagServerRequest struct {
	UUID string
	Tags []string
}

// RequestURL implements the Request interface
func (r *UntagServerRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/untag/%s", r.UUID, strings.Join(r.Tags, ","))
}
