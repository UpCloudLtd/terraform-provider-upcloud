package request

import (
	"encoding/xml"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestGetIPAddressDetailsRequest tests that GetIPAddressDetailsRequest behaves correctly
func TestGetIPAddressDetailsRequest(t *testing.T) {
	request := GetIPAddressDetailsRequest{
		Address: "0.0.0.0",
	}

	assert.Equal(t, "/ip_address/0.0.0.0", request.RequestURL())
}

// TestMarshalAssignIPAddressRequest tests that AssignIPAddressRequest structs are marshaled correctly
func TestMarshalAssignIPAddressRequest(t *testing.T) {
	request := AssignIPAddressRequest{
		Access:     upcloud.IPAddressAccessPublic,
		Family:     upcloud.IPAddressFamilyIPv4,
		ServerUUID: "009d64ef-31d1-4684-a26b-c86c955cbf46",
	}

	byteXml, err := xml.Marshal(&request)
	assert.Nil(t, err)
	expectedXML := "<ip_address><access>public</access><family>IPv4</family><server>009d64ef-31d1-4684-a26b-c86c955cbf46</server></ip_address>"
	actualXML := string(byteXml)
	assert.Equal(t, expectedXML, actualXML)

	// Omit family
	request = AssignIPAddressRequest{
		Access:     upcloud.IPAddressAccessPublic,
		ServerUUID: "009d64ef-31d1-4684-a26b-c86c955cbf46",
	}

	byteXml, err = xml.Marshal(&request)
	assert.Nil(t, err)

	expectedXML = "<ip_address><access>public</access><server>009d64ef-31d1-4684-a26b-c86c955cbf46</server></ip_address>"
	actualXML = string(byteXml)
	assert.Equal(t, expectedXML, actualXML)

}

// TestModifyIPAddressRequest tests that ModifyIPAddressRequest structs are marshaled correctly and that their URLs
// are correct
func TestModifyIPAddressRequest(t *testing.T) {
	request := ModifyIPAddressRequest{
		IPAddress: "0.0.0.0",
		PTRRecord: "ptr.example.com",
	}

	byteXml, err := xml.Marshal(&request)
	assert.Nil(t, err)
	expectedXML := "<ip_address><ptr_record>ptr.example.com</ptr_record></ip_address>"
	actualXML := string(byteXml)

	assert.Equal(t, expectedXML, actualXML)
	assert.Equal(t, "/ip_address/0.0.0.0", request.RequestURL())
}

// TestReleaseIPAddressRequest tests that ReleaseIPAddressRequest's URL is correct
func TestReleaseIPAddressRequest(t *testing.T) {
	request := ReleaseIPAddressRequest{
		IPAddress: "0.0.0.0",
	}

	assert.Equal(t, "/ip_address/0.0.0.0", request.RequestURL())
}
