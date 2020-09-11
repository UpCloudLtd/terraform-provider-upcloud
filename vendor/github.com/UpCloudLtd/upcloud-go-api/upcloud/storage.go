package upcloud

import (
	"encoding/json"
	"time"
)

// Constants
const (
	StorageTypeBackup   = "backup"
	StorageTypeCDROM    = "cdrom"
	StorageTypeDisk     = "disk"
	StorageTypeNormal   = "normal"
	StorageTypeTemplate = "template"

	StorageTierHDD     = "hdd"
	StorageTierMaxIOPS = "maxiops"

	StorageAccessPublic  = "public"
	StorageAccessPrivate = "private"

	StorageStateOnline      = "online"
	StorageStateMaintenance = "maintenance"
	StorageStateCloning     = "cloning"
	StorageStateBackuping   = "backuping"
	StorageStateError       = "error"
	StorageStateSyncing     = "syncing"

	BackupRuleIntervalDaily     = "daily"
	BackupRuleIntervalMonday    = "mon"
	BackupRuleIntervalTuesday   = "tue"
	BackupRuleIntervalWednesday = "wed"
	BackupRuleIntervalThursday  = "thu"
	BackupRuleIntervalFriday    = "fri"
	BackupRuleIntervalSaturday  = "sat"
	BackupRuleIntervalSunday    = "sun"

	StorageImportSourceDirectUpload = "direct_upload"
	StorageImportSourceHTTPImport   = "http_import"

	StorageImportStatePrepared   = "prepared"
	StorageImportStatePending    = "pending"
	StorageImportStateImporting  = "importing"
	StorageImportStateFailed     = "failed"
	StorageImportStateCancelling = "cancelling"
	StorageImportStateCancelled  = "cancelled"
	StorageImportStateCompleted  = "completed"
)

// Storages represents a /storage response
type Storages struct {
	Storages []Storage `json:"storages"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *Storages) UnmarshalJSON(b []byte) error {
	type storageWrapper struct {
		Storages []Storage `json:"storage"`
	}

	v := struct {
		Storages storageWrapper `json:"storages"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	s.Storages = v.Storages.Storages

	return nil
}

// Storage represents a storage device
type Storage struct {
	Access  string  `json:"access"`
	License float64 `json:"license"`
	// TODO: Convert to boolean
	PartOfPlan string `json:"part_of_plan"`
	Size       int    `json:"size"`
	State      string `json:"state"`
	Tier       string `json:"tier"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	UUID       string `json:"uuid"`
	Zone       string `json:"zone"`
	// Only for type "backup":
	Origin  string    `json:"origin"`
	Created time.Time `json:"created"`
}

// BackupUUIDSlice is a slice of string.
// It exists to allow for a custom JSON unmarshaller.
type BackupUUIDSlice []string

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *BackupUUIDSlice) UnmarshalJSON(b []byte) error {
	v := struct {
		BackupUUIDs []string `json:"backup"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = v.BackupUUIDs

	return nil
}

// ServerUUIDSlice is a slice of string.
// It exists to allow for a custom JSON unmarshaller.
type ServerUUIDSlice []string

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *ServerUUIDSlice) UnmarshalJSON(b []byte) error {
	v := struct {
		ServerUUIDs []string `json:"server"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = v.ServerUUIDs

	return nil
}

// StorageDetails represents detailed information about a piece of storage
type StorageDetails struct {
	Storage

	BackupRule  *BackupRule     `json:"backup_rule"`
	BackupUUIDs BackupUUIDSlice `json:"backups"`
	ServerUUIDs ServerUUIDSlice `json:"servers"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *StorageDetails) UnmarshalJSON(b []byte) error {
	type localStorageDetails StorageDetails

	v := struct {
		StorageDetails localStorageDetails `json:"storage"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = StorageDetails(v.StorageDetails)

	return nil
}

// BackupRule represents a backup rule
type BackupRule struct {
	Interval string `json:"interval"`
	// Time should be in the format "hhmm", e.g. "0430"
	Time      string `json:"time"`
	Retention int    `json:"retention,string"`
}

// ServerStorageDevice represents a storage device in the context of server requests or server details
type ServerStorageDevice struct {
	Address string `json:"address"`
	// TODO: Convert to boolean
	PartOfPlan string `json:"part_of_plan"`
	UUID       string `json:"storage"`
	Size       int    `json:"storage_size"`
	Title      string `json:"storage_title"`
	Type       string `json:"type"`
	BootDisk   int    `json:"boot_disk,string"`
}

// StorageImportDetails represents the details of an ongoing or completed storge import operation.
type StorageImportDetails struct {
	ClientContentLength int       `json:"client_content_length"`
	ClientContentType   string    `json:"client_content_type"`
	Completed           string    `json:"completed"`
	Created             time.Time `json:"created"`
	DirectUploadURL     string    `json:"direct_upload_url"`
	ErrorCode           string    `json:"error_code"`
	ErrorMessage        string    `json:"error_message"`
	MD5Sum              string    `json:"md5sum"`
	ReadBytes           int       `json:"read_bytes"`
	SHA256Sum           string    `json:"sha256sum"`
	Source              string    `json:"source"`
	State               string    `json:"state"`
	UUID                string    `json:"uuid"`
	WrittenBytes        int       `json:"written_bytes"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *StorageImportDetails) UnmarshalJSON(b []byte) error {
	type localStorageImport StorageImportDetails

	v := struct {
		StorageImport localStorageImport `json:"storage_import"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = StorageImportDetails(v.StorageImport)

	return nil
}
