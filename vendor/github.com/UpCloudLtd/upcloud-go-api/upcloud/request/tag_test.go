package request

import (
	"encoding/xml"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestCreateTagRequest tests that CreateTagRequest behaves correctly
func TestCreateTagRequest(t *testing.T) {
	request := CreateTagRequest{
		Tag: upcloud.Tag{
			Name:        "foo",
			Description: "bar",
			Servers: []string{
				"server1",
				"server2",
			},
		},
	}

	// Check the request URL
	assert.Equal(t, "/tag", request.RequestURL())

	// Check marshaling
	byteXML, err := xml.Marshal(&request)
	assert.Nil(t, err)

	expectedXML := "<tag><name>foo</name><description>bar</description><servers><server>server1</server><server>server2</server></servers></tag>"
	actualXML := string(byteXML)
	assert.Equal(t, expectedXML, actualXML)

	// Test with omitted elements
	request = CreateTagRequest{
		Tag: upcloud.Tag{
			Name: "foo",
		},
	}

	byteXML, err = xml.Marshal(&request)
	assert.Nil(t, err)

	expectedXML = "<tag><name>foo</name><servers></servers></tag>"
	actualXML = string(byteXML)
	assert.Equal(t, expectedXML, actualXML)
}

// TestModifyTagRequest tests that ModifyTagRequest behaves correctly
func TestModifyTagRequest(t *testing.T) {
	request := ModifyTagRequest{
		Name: "foo",
		Tag: upcloud.Tag{
			Name: "bar",
		},
	}

	// Check the request URL
	assert.Equal(t, "/tag/foo", request.RequestURL())

	// Check marshaling
	byteXML, err := xml.Marshal(&request)
	assert.Nil(t, err)

	expectedXML := "<tag><name>bar</name><servers></servers></tag>"
	actualXML := string(byteXML)
	assert.Equal(t, expectedXML, actualXML)
}

// TestDeleteTagRequest tests that DeleteTagRequest behaves correctly
func TestDeleteTagRequest(t *testing.T) {
	request := DeleteTagRequest{
		Name: "foo",
	}

	// Check the request URL
	assert.Equal(t, "/tag/foo", request.RequestURL())
}
