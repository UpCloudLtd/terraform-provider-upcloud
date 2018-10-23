package upcloud

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUnmarshalIPAddresses tests that IPAddresses and IPAddress structs are unmarshaled correctly
func TestUnmarshalIPAddresses(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="utf-8"?>
<ip_addresses>
    <ip_address>
        <access>public</access>
        <address>2a04:3540:1000:310:6069:7bff:fe96:71d1</address>
        <family>IPv6</family>
        <ptr_record>6069-7bff-fe96-71d1.v6.fi-hel1.host.upcloud.com</ptr_record>
        <server>0000acd1-06aa-4453-865c-fedb3f02a6a1</server>
    </ip_address>
    <ip_address>
        <access>private</access>
        <address>10.1.1.15</address>
        <family>IPv4</family>
        <ptr_record></ptr_record>
        <server>0000acd1-06aa-4453-865c-fedb3f02a6a1</server>
    </ip_address>
    <ip_address>
        <access>public</access>
        <address>94.237.32.40</address>
        <family>IPv4</family>
        <part_of_plan>yes</part_of_plan>
        <ptr_record>94-237-32-40.fi-hel1.host.upcloud.com</ptr_record>
        <server>0000acd1-06aa-4453-865c-fedb3f02a6a1</server>
    </ip_address>
</ip_addresses>`

	ipAddresses := IPAddresses{}
	err := xml.Unmarshal([]byte(originalXML), &ipAddresses)
	assert.Nil(t, err)
	assert.Len(t, ipAddresses.IPAddresses, 3)

	firstAddress := ipAddresses.IPAddresses[0]
	assert.Equal(t, IPAddressAccessPublic, firstAddress.Access)
	assert.Equal(t, "2a04:3540:1000:310:6069:7bff:fe96:71d1", firstAddress.Address)
	assert.Equal(t, IPAddressFamilyIPv6, firstAddress.Family)
	assert.Equal(t, "6069-7bff-fe96-71d1.v6.fi-hel1.host.upcloud.com", firstAddress.PTRRecord)
	assert.Equal(t, "0000acd1-06aa-4453-865c-fedb3f02a6a1", firstAddress.ServerUUID)
}
