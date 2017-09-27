package upcloud

import "encoding/xml"

// Constants
const (
	StorageTypeDisk     = "disk"
	StorageTypeCDROM    = "cdrom"
	StorageTypeTemplate = "template"
	StorageTypeBackup   = "backup"

	StorageTierHDD     = "hdd"
	StorageTierMaxIOPS = "maxiops"

	StorageAccessPublic  = "public"
	StorageAccessPrivate = "private"

	StorageStateOnline      = "online"
	StorageStateMaintenance = "maintenance"
	StorageStateCloning     = "cloning"
	StorageStateBackuping   = "backuping"
	StorageStateError       = "error"

	BackupRuleIntervalDaily     = "daily"
	BackupRuleIntervalMonday    = "mon"
	BackupRuleIntervalTuesday   = "tue"
	BackupRuleIntervalWednesday = "wed"
	BackupRuleIntervalThursday  = "thu"
	BackupRuleIntervalFriday    = "fri"
	BackupRuleIntervalSaturday  = "sat"
	BackupRuleIntervalSunday    = "sun"

	CreateServerStorageDeviceActionCreate = "create"
	CreateServerStorageDeviceActionClone  = "clone"
	CreateServerStorageDeviceActionAttach = "attach"
)

// Storages represents a /storage response
type Storages struct {
	Storages []Storage `xml:"storage"`
}

// Storage represents a storage device
type Storage struct {
	Access  string  `xml:"access"`
	License float64 `xml:"license"`
	// TODO: Convert to boolean
	PartOfPlan string `xml:"part_of_plan"`
	Size       int    `xml:"size"`
	State      string `xml:"state"`
	Tier       string `xml:"tier"`
	Title      string `xml:"title"`
	Type       string `xml:"type"`
	UUID       string `xml:"uuid"`
	Zone       string `xml:"zone"`
}

// StorageDetails represents detailed information about a piece of storage
type StorageDetails struct {
	Storage

	BackupRule *BackupRule `xml:"backup_rule"`
	// TODO: Support the <backups> field
	ServerUUIDs []string `xml:"servers>server"`
}

// BackupRule represents a backup rule
type BackupRule struct {
	XMLName  xml.Name `xml:"backup_rule"`
	Interval string   `xml:"interval"`
	// Time should be in the format "hhmm", e.g. "0430"
	Time      string `xml:"time"`
	Retention int    `xml:"retention"`
}

// ServerStorageDevice represents a storage device in the context of server requests or server details
type ServerStorageDevice struct {
	XMLName xml.Name `xml:"storage_device"`

	Address string `xml:"address"`
	// TODO: Convert to boolean
	PartOfPlan string `xml:"part_of_plan"`
	UUID       string `xml:"storage"`
	Size       int    `xml:"storage_size"`
	Title      string `xml:"storage_title"`
	Type       string `xml:"type"`
}

// CreateServerStorageDevice represents a storage device for a CreateServerRequest
type CreateServerStorageDevice struct {
	XMLName xml.Name `xml:"storage_device"`

	Action  string `xml:"action"`
	Address string `xml:"address,omitempty"`
	Storage string `xml:"storage"`
	Title   string `xml:"title,omitempty"`
	// Storage size in gigabytes
	Size int    `xml:"size"`
	Tier string `xml:"tier,omitempty"`
	Type string `xml:"type,omitempty"`
}
