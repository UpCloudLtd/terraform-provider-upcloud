package upcloud

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUnmarshalStorage tests that Storages and Storage struct are unmarshaled correctly
func TestUnmarshalStorage(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="utf-8"?>
<storages>
    <storage>
        <access>public</access>
        <license>0</license>
        <size>1</size>
        <state>online</state>
        <title>Windows Server 2003 R2 Standard (CD 1)</title>
        <type>cdrom</type>
        <uuid>01000000-0000-4000-8000-000010010101</uuid>
    </storage>
    <storage>
        <access>public</access>
        <license>0</license>
        <size>1</size>
        <state>online</state>
        <title>Windows Server 2003 R2 Standard (CD 2)</title>
        <type>cdrom</type>
        <uuid>01000000-0000-4000-8000-000010010102</uuid>
    </storage>
    <storage>
        <access>public</access>
        <license>0</license>
        <size>1</size>
        <state>online</state>
        <title>Windows Server 2003 R2 Standard (CD 1)</title>
        <type>cdrom</type>
        <uuid>01000000-0000-4000-8000-000010010201</uuid>
    </storage>
</storages>`

	storages := Storages{}
	err := xml.Unmarshal([]byte(originalXML), &storages)

	assert.Nil(t, err)
	assert.Len(t, storages.Storages, 3)

	firstStorage := storages.Storages[0]
	assert.Equal(t, "public", firstStorage.Access)
	assert.Equal(t, 0.0, firstStorage.License)
	assert.Equal(t, 1, firstStorage.Size)
	assert.Equal(t, "Windows Server 2003 R2 Standard (CD 1)", firstStorage.Title)
	assert.Equal(t, StorageTypeCDROM, firstStorage.Type)
	assert.Equal(t, "01000000-0000-4000-8000-000010010101", firstStorage.UUID)
}

// TestUnmarshalServerStorageDevice tests that ServerStorageDevice objects are properly unmarshaled
func TestUnmarshalServerStorageDevice(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="utf-8"?>
<storage_device>
    <address>virtio:0</address>
    <part_of_plan>yes</part_of_plan>
    <storage>01c8df16-d1c6-4223-9bfc-d3c06b208c88</storage>
    <storage_size>30</storage_size>
    <storage_title>test-disk0</storage_title>
    <type>disk</type>
</storage_device>`

	storageDevice := ServerStorageDevice{}
	err := xml.Unmarshal([]byte(originalXML), &storageDevice)

	assert.Nil(t, err)
	assert.Equal(t, "virtio:0", storageDevice.Address)
	assert.Equal(t, "yes", storageDevice.PartOfPlan)
	assert.Equal(t, "01c8df16-d1c6-4223-9bfc-d3c06b208c88", storageDevice.UUID)
	assert.Equal(t, 30, storageDevice.Size)
	assert.Equal(t, "test-disk0", storageDevice.Title)
	assert.Equal(t, StorageTypeDisk, storageDevice.Type)
}

// TestMarshalCreateStorageDevice tests that CreateStorageDevice objects are correctly marshaled. We don't need to
// test unmarshaling because these data structures are never returned from the API.
func TestMarshalCreateStorageDevice(t *testing.T) {
	storage := CreateServerStorageDevice{
		Action:  CreateServerStorageDeviceActionClone,
		Storage: "01000000-0000-4000-8000-000030060200",
		Title:   "disk1",
		Size:    30,
		Tier:    StorageTierMaxIOPS,
	}

	expectedXML := "<storage_device><action>clone</action><storage>01000000-0000-4000-8000-000030060200</storage><title>disk1</title><size>30</size><tier>maxiops</tier></storage_device>"

	actualXML, err := xml.Marshal(storage)
	assert.Nil(t, err)
	assert.Equal(t, expectedXML, string(actualXML))
}

// TestMarshalBackupRule tests that BackupRule objects are properly marshaled
func TestMarshalBackupRule(t *testing.T) {
	backupRule := BackupRule{
		Interval:  BackupRuleIntervalDaily,
		Time:      "0430",
		Retention: 30,
	}

	ruleXML, err := xml.Marshal(backupRule)
	assert.Nil(t, err)
	assert.Equal(t, "<backup_rule><interval>daily</interval><time>0430</time><retention>30</retention></backup_rule>", string(ruleXML))
}

// TestUnmarshalBackupRule tests that BackupRule objects are properly unmarshaled
func TestUnmarshalBackupRule(t *testing.T) {
	originalXML := "<backup_rule><interval>daily</interval><time>0430</time><retention>30</retention></backup_rule>"

	backupRule := BackupRule{}
	err := xml.Unmarshal([]byte(originalXML), &backupRule)
	assert.Nil(t, err)
	assert.Equal(t, BackupRuleIntervalDaily, backupRule.Interval)
	assert.Equal(t, "0430", backupRule.Time)
	assert.Equal(t, 30, backupRule.Retention)
}
