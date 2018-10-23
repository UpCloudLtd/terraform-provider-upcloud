package request

import (
	"encoding/xml"
	"fmt"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"strings"
	"time"
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
)

// GetServerDetailsRequest represents a request for retrieving details about a server
type GetServerDetailsRequest struct {
	UUID string
}

// RequestURL implements the Request interface
func (r *GetServerDetailsRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s", r.UUID)
}

// CreateServerRequest represents a request for creating a new server
type CreateServerRequest struct {
	XMLName xml.Name `xml:"server"`

	AvoidHost  string `xml:"avoid_host,omitempty"`
	BootOrder  string `xml:"boot_order,omitempty"`
	CoreNumber int    `xml:"core_number,omitempty"`
	// TODO: Convert to boolean
	Firewall         string                              `xml:"firewall,omitempty"`
	Hostname         string                              `xml:"hostname"`
	IPAddresses      []CreateServerIPAddress             `xml:"ip_addresses>ip_address"`
	LoginUser        *LoginUser                          `xml:"login_user,omitempty"`
	MemoryAmount     int                                 `xml:"memory_amount,omitempty"`
	PasswordDelivery string                              `xml:"password_delivery,omitempty"`
	Plan             string                              `xml:"plan,omitempty"`
	StorageDevices   []upcloud.CreateServerStorageDevice `xml:"storage_devices>storage_device"`
	TimeZone         string                              `xml:"timezone,omitempty"`
	Title            string                              `xml:"title"`
	UserData         string                              `xml:"user_data,omitempty"`
	VideoModel       string                              `xml:"video_model,omitempty"`
	// TODO: Convert to boolean
	VNC         string `xml:"vnc,omitempty"`
	VNCPassword string `xml:"vnc_password,omitempty"`
	Zone        string `xml:"zone"`
}

// RequestURL implements the Request interface
func (r *CreateServerRequest) RequestURL() string {
	return "/server"
}

// LoginUser represents the login_user block when creating a new server
type LoginUser struct {
	CreatePassword string   `xml:"create_password,omitempty"`
	Username       string   `xml:"username,omitempty"`
	SSHKeys        []string `xml:"ssh_keys>ssh_key,omitempty"`
}

// CreateServerIPAddress represents an IP address for a CreateServerRequest
type CreateServerIPAddress struct {
	Access string `xml:"access"`
	Family string `xml:"family"`
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
	UUID string

	// TODO: Start server requests have no timeout in the API
	Timeout time.Duration
}

// RequestURL implements the Request interface
func (r *StartServerRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/start", r.UUID)
}

// StopServerRequest represents a request to stop a server
type StopServerRequest struct {
	XMLName xml.Name `xml:"stop_server"`

	UUID string `xml:"-"`

	StopType string        `xml:"stop_type,omitempty"`
	Timeout  time.Duration `xml:"timeout,omitempty"`
}

// RequestURL implements the Request interface
func (r *StopServerRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/stop", r.UUID)
}

// MarshalXML implements a custom marshaller for StopServerRequest which converts the timeout to seconds
func (r *StopServerRequest) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type Alias StopServerRequest

	return e.Encode(&struct {
		Timeout int `xml:"timeout,omitempty"`
		*Alias
	}{
		Timeout: int(r.Timeout.Seconds()),
		Alias:   (*Alias)(r),
	})
}

// RestartServerRequest represents a request to restart a server
type RestartServerRequest struct {
	XMLName xml.Name `xml:"restart_server"`

	UUID string `xml:"-"`

	StopType      string        `xml:"stop_type,omitempty"`
	Timeout       time.Duration `xml:"timeout,omitempty"`
	TimeoutAction string        `xml:"timeout_action,omitempty"`
}

// RequestURL implements the Request interface
func (r *RestartServerRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/restart", r.UUID)
}

// MarshalXML implements a custom marshaller for RestartServerRequest which converts the timeout to seconds
func (r *RestartServerRequest) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type Alias RestartServerRequest

	return e.Encode(&struct {
		Timeout int `xml:"timeout,omitempty"`
		*Alias
	}{
		Timeout: int(r.Timeout.Seconds()),
		Alias:   (*Alias)(r),
	})
}

// ModifyServerRequest represents a request to modify a server
type ModifyServerRequest struct {
	XMLName xml.Name `xml:"server"`

	UUID string `xml:"-"`

	AvoidHost string `xml:"avoid_host,omitempty"`
	BootOrder string `xml:"boot_order,omitempty"`
	// TODO: Investigate correct type and format
	CoreNumber string `xml:"core_number,omitempty"`
	// TODO: Convert to boolean
	Firewall string `xml:"firewall,omitempty"`
	Hostname string `xml:"hostname,omitempty"`
	// TODO: Investigate correct type and format
	MemoryAmount string `xml:"memory_amount,omitempty"`
	Plan         string `xml:"plan,omitempty"`
	TimeZone     string `xml:"timezone,omitempty"`
	Title        string `xml:"title,omitempty"`
	VideoModel   string `xml:"video_model,omitempty"`
	// TODO: Convert to boolean
	VNC         string `xml:"vnc,omitempty"`
	VNCPassword string `xml:"vnc_password,omitempty"`
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
