package request

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
)

// Constants
const (
	StorageImportSourceDirectUpload = "direct_upload"
	StorageImportSourceHTTPImport   = "http_import"
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
	Size       int                 `json:"size,string"`
	Tier       string              `json:"tier,omitempty"`
	Title      string              `json:"title,omitempty"`
	Zone       string              `json:"zone"`
	BackupRule *upcloud.BackupRule `json:"backup_rule,omitempty"`
}

// RequestURL implements the Request interface
func (r *CreateStorageRequest) RequestURL() string {
	return "/storage"
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r CreateStorageRequest) MarshalJSON() ([]byte, error) {
	type localCreateStorageRequest CreateStorageRequest
	v := struct {
		CreateStorageRequest localCreateStorageRequest `json:"storage"`
	}{}
	v.CreateStorageRequest = localCreateStorageRequest(r)

	return json.Marshal(&v)
}

// ModifyStorageRequest represents a request to modify a storage device
type ModifyStorageRequest struct {
	UUID string `json:"-"`

	Title      string              `json:"title,omitempty"`
	Size       int                 `json:"size,omitempty,string"`
	BackupRule *upcloud.BackupRule `json:"backup_rule,omitempty"`
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r ModifyStorageRequest) MarshalJSON() ([]byte, error) {
	type localModifyStorageRequest ModifyStorageRequest
	v := struct {
		ModifyStorageRequest localModifyStorageRequest `json:"storage"`
	}{}
	v.ModifyStorageRequest = localModifyStorageRequest(r)

	return json.Marshal(&v)
}

// RequestURL implements the Request interface
func (r *ModifyStorageRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s", r.UUID)
}

// AttachStorageRequest represents a request to attach a storage device to a server
type AttachStorageRequest struct {
	ServerUUID string `json:"-"`

	Type        string `json:"type,omitempty"`
	Address     string `json:"address,omitempty"`
	StorageUUID string `json:"storage,omitempty"`
	BootDisk    int    `json:"boot_disk,omitempty,string"`
}

// RequestURL implements the Request interface
func (r *AttachStorageRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/storage/attach", r.ServerUUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r AttachStorageRequest) MarshalJSON() ([]byte, error) {
	type localAttachStorageRequest AttachStorageRequest
	v := struct {
		AttachStorageRequest localAttachStorageRequest `json:"storage_device"`
	}{}
	v.AttachStorageRequest = localAttachStorageRequest(r)

	return json.Marshal(&v)
}

// DetachStorageRequest represents a request to detach a storage device from a server
type DetachStorageRequest struct {
	ServerUUID string `json:"-"`

	Address string `json:"address"`
}

// RequestURL implements the Request interface
func (r *DetachStorageRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/storage/detach", r.ServerUUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r DetachStorageRequest) MarshalJSON() ([]byte, error) {
	type localDetachStorageRequest DetachStorageRequest
	v := struct {
		DetachStorageRequest localDetachStorageRequest `json:"storage_device"`
	}{}
	v.DetachStorageRequest = localDetachStorageRequest(r)

	return json.Marshal(&v)
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
	UUID string `json:"-"`

	Zone  string `json:"zone"`
	Tier  string `json:"tier,omitempty"`
	Title string `json:"title"`
}

// RequestURL implements the Request interface
func (r *CloneStorageRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/clone", r.UUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r CloneStorageRequest) MarshalJSON() ([]byte, error) {
	type localCloneStorageRequest CloneStorageRequest
	v := struct {
		CloneStorageRequest localCloneStorageRequest `json:"storage"`
	}{}
	v.CloneStorageRequest = localCloneStorageRequest(r)

	return json.Marshal(&v)
}

// TemplatizeStorageRequest represents a request to templatize a storage device
type TemplatizeStorageRequest struct {
	UUID string `json:"-"`

	Title string `json:"title"`
}

// RequestURL implements the Request interface
func (r *TemplatizeStorageRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/templatize", r.UUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r TemplatizeStorageRequest) MarshalJSON() ([]byte, error) {
	type localTemplatizeStorageRequest TemplatizeStorageRequest
	v := struct {
		TemplatizeStorageRequest localTemplatizeStorageRequest `json:"storage"`
	}{}
	v.TemplatizeStorageRequest = localTemplatizeStorageRequest(r)

	return json.Marshal(&v)
}

// WaitForStorageStateRequest represents a request to wait for a storage to enter a specific state
type WaitForStorageStateRequest struct {
	UUID         string
	DesiredState string
	Timeout      time.Duration
}

// LoadCDROMRequest represents a request to load a storage as a CD-ROM in the CD-ROM device of a server
type LoadCDROMRequest struct {
	ServerUUID string `json:"-"`

	StorageUUID string `json:"storage"`
}

// RequestURL implements the Request interface
func (r *LoadCDROMRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/cdrom/load", r.ServerUUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r LoadCDROMRequest) MarshalJSON() ([]byte, error) {
	type localLoadCDROMRequest LoadCDROMRequest
	v := struct {
		LoadCDROMRequest localLoadCDROMRequest `json:"storage_device"`
	}{}
	v.LoadCDROMRequest = localLoadCDROMRequest(r)

	return json.Marshal(&v)
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
	UUID string `json:"-"`

	Title string `json:"title"`
}

// RequestURL implements the Request interface
func (r *CreateBackupRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/backup", r.UUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r CreateBackupRequest) MarshalJSON() ([]byte, error) {
	type localCreateBackupRequest CreateBackupRequest
	v := struct {
		CreateBackupRequest localCreateBackupRequest `json:"storage"`
	}{}
	v.CreateBackupRequest = localCreateBackupRequest(r)

	return json.Marshal(&v)
}

// RestoreBackupRequest represents a request to restore a storage from the specified backup
type RestoreBackupRequest struct {
	UUID string
}

// RequestURL implements the Request interface
func (r *RestoreBackupRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/restore", r.UUID)
}

// CreateStorageImportRequest represent a request to import storage.
type CreateStorageImportRequest struct {
	StorageUUID string `json:"-"`

	Source         string `json:"source"`
	SourceLocation string `json:"source_location,omitempty"`
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r CreateStorageImportRequest) MarshalJSON() ([]byte, error) {
	type localStorageImportRequest CreateStorageImportRequest
	v := struct {
		StorageImportRequest localStorageImportRequest `json:"storage_import"`
	}{}
	v.StorageImportRequest = localStorageImportRequest(r)

	return json.Marshal(&v)
}

// RequestURL implements the Request interface
func (r *CreateStorageImportRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/import", r.StorageUUID)
}

// GetStorageImportDetailsRequest represents a request to get details
// about an import
type GetStorageImportDetailsRequest struct {
	UUID string
}

// RequestURL implements the Request interface
func (r *GetStorageImportDetailsRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/import", r.UUID)
}

// WaitForStorageImportCompletionRequest represents a request to wait
// for storage import to complete.
type WaitForStorageImportCompletionRequest struct {
	StorageUUID string
	Timeout     time.Duration
}
