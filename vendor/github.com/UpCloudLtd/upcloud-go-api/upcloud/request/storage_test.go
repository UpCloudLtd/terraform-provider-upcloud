package request

import (
	"encoding/xml"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestGetStoragesRequest tests that GetStoragesRequest objects behave correctly
func TestGetStoragesRequest(t *testing.T) {
	request := GetStoragesRequest{}

	assert.Equal(t, "/storage", request.RequestURL())
	request.Access = upcloud.StorageAccessPublic
	assert.Equal(t, "/storage/public", request.RequestURL())
	request.Access = ""
	request.Favorite = true
	assert.Equal(t, "/storage/favorite", request.RequestURL())
	request.Favorite = false
	request.Type = upcloud.StorageTypeDisk
	assert.Equal(t, "/storage/disk", request.RequestURL())
}

// TestGetStorageDetailsRequest tests that GetStorageDetailsRequest objects behave correctly
func TestGetStorageDetailsRequest(t *testing.T) {
	request := GetStorageDetailsRequest{
		UUID: "foo",
	}

	assert.Equal(t, "/storage/foo", request.RequestURL())
}

// TestCreateStorageRequest tests that CreateStorageRequest objects behave correctly
func TestCreateStorageRequest(t *testing.T) {
	request := CreateStorageRequest{
		Tier:  upcloud.StorageTierMaxIOPS,
		Title: "Test storage",
		Size:  10,
		Zone:  "fi-hel1",
		BackupRule: &upcloud.BackupRule{
			Interval:  upcloud.BackupRuleIntervalDaily,
			Time:      "0430",
			Retention: 30,
		},
	}

	expectedXML := "<storage><size>10</size><tier>maxiops</tier><title>Test storage</title><zone>fi-hel1</zone><backup_rule><interval>daily</interval><time>0430</time><retention>30</retention></backup_rule></storage>"
	actualXML, err := xml.Marshal(&request)
	assert.Nil(t, err)
	assert.Equal(t, expectedXML, string(actualXML))
	assert.Equal(t, "/storage", request.RequestURL())
}

// TestModifyStorageRequest tests that ModifyStorageRequest objects behave correctly
func TestModifyStorageRequest(t *testing.T) {
	request := ModifyStorageRequest{
		UUID:  "foo",
		Title: "New fancy title",
	}

	expectedXML := "<storage><title>New fancy title</title></storage>"
	actualXML, err := xml.Marshal(&request)
	assert.Nil(t, err)
	assert.Equal(t, expectedXML, string(actualXML))
	assert.Equal(t, "/storage/foo", request.RequestURL())
}

// TestAttachStorageRequest tests that AttachStorageRequest objects behave correctly
func TestAttachStorageRequest(t *testing.T) {
	request := AttachStorageRequest{
		StorageUUID: "foo",
		ServerUUID:  "bar",
		Type:        upcloud.StorageTypeDisk,
		Address:     "scsi:0:0",
	}

	expectedXML := "<storage_device><type>disk</type><address>scsi:0:0</address><storage>foo</storage></storage_device>"
	actualXML, err := xml.Marshal(&request)
	assert.Nil(t, err)
	assert.Equal(t, expectedXML, string(actualXML))
	assert.Equal(t, "/server/bar/storage/attach", request.RequestURL())
}

// TestDetachStorageRequest tests that DetachStorageRequest objects behave correctly
func TestDetachStorageRequest(t *testing.T) {
	request := DetachStorageRequest{
		ServerUUID: "foo",
		Address:    "scsi:0:0",
	}

	expectedXML := "<storage_device><address>scsi:0:0</address></storage_device>"
	actualXML, err := xml.Marshal(&request)
	assert.Nil(t, err)
	assert.Equal(t, expectedXML, string(actualXML))
	assert.Equal(t, "/server/foo/storage/detach", request.RequestURL())
}

// TestDeleteStorageRequest tests that DeleteStorageRequest objects behave correctly
func TestDeleteStorageRequest(t *testing.T) {
	request := DeleteStorageRequest{
		UUID: "foo",
	}

	assert.Equal(t, "/storage/foo", request.RequestURL())
}

// TestCloneStorageRequest testa that CloneStorageRequest objects behave correctly
func TestCloneStorageRequest(t *testing.T) {
	request := CloneStorageRequest{
		UUID:  "foo",
		Title: "Cloned storage",
		Zone:  "fi-hel1",
		Tier:  upcloud.StorageTierMaxIOPS,
	}

	expectedXML := "<storage><zone>fi-hel1</zone><tier>maxiops</tier><title>Cloned storage</title></storage>"
	actualXML, err := xml.Marshal(&request)
	assert.Nil(t, err)
	assert.Equal(t, expectedXML, string(actualXML))
	assert.Equal(t, "/storage/foo/clone", request.RequestURL())
}

// TestTemplatizeStorageRequest tests that TemplatizeStorageRequest objects behave correctly
func TestTemplatizeStorageRequest(t *testing.T) {
	request := TemplatizeStorageRequest{
		UUID:  "foo",
		Title: "Templatized storage",
	}

	expectedXML := "<storage><title>Templatized storage</title></storage>"
	actualXML, err := xml.Marshal(&request)
	assert.Nil(t, err)
	assert.Equal(t, expectedXML, string(actualXML))
	assert.Equal(t, "/storage/foo/templatize", request.RequestURL())
}

// TestLoadCDROMRequest tests that LoadCDROMRequest objects behave correctly
func TestLoadCDROMRequest(t *testing.T) {
	request := LoadCDROMRequest{
		ServerUUID:  "foo",
		StorageUUID: "bar",
	}

	expectedXML := "<storage_device><storage>bar</storage></storage_device>"
	actualXML, err := xml.Marshal(&request)
	assert.Nil(t, err)
	assert.Equal(t, expectedXML, string(actualXML))
	assert.Equal(t, "/server/foo/cdrom/load", request.RequestURL())
}

// TestEjectCDROMRequest tests that EjectCDROMRequest objects behave correctly
func TestEjectCDROMRequest(t *testing.T) {
	request := EjectCDROMRequest{
		ServerUUID: "foo",
	}

	assert.Equal(t, "/server/foo/cdrom/eject", request.RequestURL())
}

// TestCreateBackupRequest tests that CreateBackupRequest objects behave correctly
func TestCreateBackupRequest(t *testing.T) {
	request := CreateBackupRequest{
		UUID:  "foo",
		Title: "Backup",
	}

	expectedXML := "<storage><title>Backup</title></storage>"
	actualXML, err := xml.Marshal(&request)
	assert.Nil(t, err)
	assert.Equal(t, expectedXML, string(actualXML))
	assert.Equal(t, "/storage/foo/backup", request.RequestURL())
}

// TestRestoreBackupRequest tests that RestoreBackupRequest objects behave correctly
func TestRestoreBackupRequest(t *testing.T) {
	request := RestoreBackupRequest{
		UUID: "foo",
	}

	assert.Equal(t, "/storage/foo/restore", request.RequestURL())
}
