package request

import (
	"encoding/xml"
	"fmt"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"time"
)

// GetStoragesRequest represents a request for retrieving all or some storages
type GetStoragesRequest struct {
	// If specified, only storages with this access type will be retrieved
	Access string
	// If specified, only storages with this type will be retrieved
	Type string
	// If specified, only storages marked as favorite will be retrieved
	Favorite bool
}

// RequestURL implements the Request interface
func (r *GetStoragesRequest) RequestURL() string {
	if r.Access != "" {
		return fmt.Sprintf("/storage/%s", r.Access)
	}

	if r.Type != "" {
		return fmt.Sprintf("/storage/%s", r.Type)
	}

	if r.Favorite {
		return "/storage/favorite"
	}

	return "/storage"
}

// GetStorageDetailsRequest represents a request for retrieving details about a piece of storage
type GetStorageDetailsRequest struct {
	UUID string
}

// RequestURL implements the Request interface
func (r *GetStorageDetailsRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s", r.UUID)
}

// CreateStorageRequest represents a request to create a storage device
type CreateStorageRequest struct {
	XMLName xml.Name `xml:"storage"`

	Size       int                 `xml:"size"`
	Tier       string              `xml:"tier,omitempty"`
	Title      string              `xml:"title"`
	Zone       string              `xml:"zone"`
	BackupRule *upcloud.BackupRule `xml:"backup_rule,omitempty"`
}

// RequestURL implements the Request interface
func (r *CreateStorageRequest) RequestURL() string {
	return "/storage"
}

// ModifyStorageRequest represents a request to modify a storage device
type ModifyStorageRequest struct {
	XMLName xml.Name `xml:"storage"`
	UUID    string   `xml:"-"`

	Title      string              `xml:"title,omitempty"`
	Size       int                 `xml:"size,omitempty"`
	BackupRule *upcloud.BackupRule `xml:"backup_rule,omitempty"`
}

// RequestURL implements the Request interface
func (r *ModifyStorageRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s", r.UUID)
}

// AttachStorageRequest represents a request to attach a storage device to a server
type AttachStorageRequest struct {
	XMLName    xml.Name `xml:"storage_device"`
	ServerUUID string   `xml:"-"`

	Type        string `xml:"type,omitempty"`
	Address     string `xml:"address,omitempty"`
	StorageUUID string `xml:"storage,omitempty"`
}

// RequestURL implements the Request interface
func (r *AttachStorageRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/storage/attach", r.ServerUUID)
}

// DetachStorageRequest represents a request to detach a storage device from a server
type DetachStorageRequest struct {
	XMLName    xml.Name `xml:"storage_device"`
	ServerUUID string   `xml:"-"`

	Address string `xml:"address"`
}

// RequestURL implements the Request interface
func (r *DetachStorageRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/storage/detach", r.ServerUUID)
}

//DeleteStorageRequest represents a request to delete a storage device
type DeleteStorageRequest struct {
	UUID string
}

// RequestURL implements the Request interface
func (r *DeleteStorageRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s", r.UUID)
}

// CloneStorageRequest represents a requests to clone a storage device
type CloneStorageRequest struct {
	XMLName xml.Name `xml:"storage"`
	UUID    string   `xml:"-"`

	Zone  string `xml:"zone"`
	Tier  string `xml:"tier,omitempty"`
	Title string `xml:"title"`
}

// RequestURL implements the Request interface
func (r *CloneStorageRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/clone", r.UUID)
}

// TemplatizeStorageRequest represents a request to templatize a storage device
type TemplatizeStorageRequest struct {
	XMLName xml.Name `xml:"storage"`
	UUID    string   `xml:"-"`

	Title string `xml:"title"`
}

// RequestURL implements the Request interface
func (r *TemplatizeStorageRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/templatize", r.UUID)
}

// WaitForStorageStateRequest represents a request to wait for a storage to enter a specific state
type WaitForStorageStateRequest struct {
	UUID         string
	DesiredState string
	Timeout      time.Duration
}

// LoadCDROMRequest represents a request to load a storage as a CD-ROM in the CD-ROM device of a server
type LoadCDROMRequest struct {
	XMLName    xml.Name `xml:"storage_device"`
	ServerUUID string   `xml:"-"`

	StorageUUID string `xml:"storage"`
}

// RequestURL implements the Request interface
func (r *LoadCDROMRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/cdrom/load", r.ServerUUID)
}

// EjectCDROMRequest represents a request to load a storage as a CD-ROM in the CD-ROM device of a server
type EjectCDROMRequest struct {
	ServerUUID string
}

// RequestURL implements the Request interface
func (r *EjectCDROMRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/cdrom/eject", r.ServerUUID)
}

// CreateBackupRequest represents a request to create a backup of a storage device
type CreateBackupRequest struct {
	XMLName xml.Name `xml:"storage"`
	UUID    string   `xml:"-"`

	Title string `xml:"title"`
}

// RequestURL implements the Request interface
func (r *CreateBackupRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/backup", r.UUID)
}

// RestoreBackupRequest represents a request to restore a storage from the specified backup
type RestoreBackupRequest struct {
	UUID string
}

// RequestURL implements the Request interface
func (r *RestoreBackupRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/restore", r.UUID)
}
