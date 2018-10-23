package upcloud

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUnmarshalServerConfiguratons tests that ServerConfigurations and ServerConfiguration are unmarshaled correctly
func TestUnmarshalServerConfiguratons(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="utf-8"?>
<server_sizes>
    <server_size>
        <core_number>1</core_number>
        <memory_amount>512</memory_amount>
    </server_size>
    <server_size>
        <core_number>1</core_number>
        <memory_amount>1024</memory_amount>
    </server_size>
    <server_size>
        <core_number>1</core_number>
        <memory_amount>1536</memory_amount>
    </server_size>
</server_sizes>`

	serverConfigurations := ServerConfigurations{}
	err := xml.Unmarshal([]byte(originalXML), &serverConfigurations)
	assert.Nil(t, err)
	assert.Len(t, serverConfigurations.ServerConfigurations, 3)

	firstConfiguration := serverConfigurations.ServerConfigurations[0]
	assert.Equal(t, 1, firstConfiguration.CoreNumber)
	assert.Equal(t, 512, firstConfiguration.MemoryAmount)
}

// TestUnmarshalServers tests that Servers and Server are unmarshaled correctly
func TestUnmarshalServers(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="utf-8"?>
<servers>
    <server>
        <core_number>1</core_number>
        <hostname>foo</hostname>
        <license>0</license>
        <memory_amount>1024</memory_amount>
        <plan>1xCPU-1GB</plan>
        <progress>95</progress>
        <state>maintenance</state>
        <tags>
    </tags>
        <title>foo.example.com</title>
        <uuid>009114f1-cd89-4202-b057-5680eb6ba428</uuid>
        <zone>uk-lon1</zone>
    </server>
</servers>`

	servers := Servers{}
	err := xml.Unmarshal([]byte(originalXML), &servers)
	assert.Nil(t, err)
	assert.Len(t, servers.Servers, 1)

	server := servers.Servers[0]
	assert.Equal(t, 1, server.CoreNumber)
	assert.Equal(t, "foo", server.Hostname)
	assert.Equal(t, 0.0, server.License)
	assert.Equal(t, 1024, server.MemoryAmount)
	assert.Equal(t, "1xCPU-1GB", server.Plan)
	assert.Equal(t, 95, server.Progress)
	assert.Equal(t, ServerStateMaintenance, server.State)
	assert.Empty(t, server.Tags)
	assert.Equal(t, "foo.example.com", server.Title)
	assert.Equal(t, "009114f1-cd89-4202-b057-5680eb6ba428", server.UUID)
	assert.Equal(t, "uk-lon1", server.Zone)
}

// TestUnmarshalServerDetails tests that ServerDetails objects are correctly unmarshaled
func TestUnmarshalServerDetails(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="utf-8"?>
<server>
    <boot_order>cdrom,disk</boot_order>
    <core_number>1</core_number>
    <firewall>off</firewall>
    <hostname>foo</hostname>
    <ip_addresses>
        <ip_address>
            <access>private</access>
            <address>10.2.0.123</address>
            <family>IPv4</family>
        </ip_address>
        <ip_address>
            <access>public</access>
            <address>2a04:3541:1000:500:6069:7bff:fe96:613d</address>
            <family>IPv6</family>
        </ip_address>
        <ip_address>
            <access>public</access>
            <address>83.136.252.63</address>
            <family>IPv4</family>
            <part_of_plan>yes</part_of_plan>
        </ip_address>
    </ip_addresses>
    <license>0</license>
    <memory_amount>1024</memory_amount>
    <nic_model>virtio</nic_model>
    <plan>1xCPU-1GB</plan>
    <state>maintenance</state>
    <storage_devices>
        <storage_device>
            <address>virtio:0</address>
            <part_of_plan>yes</part_of_plan>
            <storage>01ea621d-e509-4a78-bdf2-068feb07e92a</storage>
            <storage_size>30</storage_size>
            <storage_title>foo-disk0</storage_title>
            <type>disk</type>
        </storage_device>
    </storage_devices>
    <tags>
  </tags>
    <timezone>UTC</timezone>
    <title>foo.example.com</title>
    <uuid>007efb1c-b7ba-41ff-9e03-57876f2ba0b0</uuid>
    <video_model>cirrus</video_model>
    <vnc>off</vnc>
    <vnc_password>7rNQE6mA</vnc_password>
    <zone>uk-lon1</zone>
</server>`

	serverDetails := ServerDetails{}
	err := xml.Unmarshal([]byte(originalXML), &serverDetails)
	assert.Nil(t, err)

	assert.Equal(t, "cdrom,disk", serverDetails.BootOrder)
	assert.Equal(t, "off", serverDetails.Firewall)
	assert.Len(t, serverDetails.IPAddresses, 3)
	assert.Equal(t, "virtio", serverDetails.NICModel)
	assert.Len(t, serverDetails.StorageDevices, 1)
	assert.Equal(t, "UTC", serverDetails.Timezone)
	assert.Equal(t, VideoModelCirrus, serverDetails.VideoModel)
	assert.Equal(t, "off", serverDetails.VNC)
	assert.Equal(t, "7rNQE6mA", serverDetails.VNCPassword)
}
