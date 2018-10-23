package upcloud

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUnmarshalZones tests that the Zone and Zones structs are correctly marshaled
func TestUnmarshalZones(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="utf-8"?>
<zones>
    <zone>
        <description>Frankfurt #1</description>
        <id>de-fra1</id>
    </zone>
    <zone>
        <description>Helsinki #1</description>
        <id>fi-hel1</id>
    </zone>
    <zone>
        <description>London #1</description>
        <id>uk-lon1</id>
    </zone>
    <zone>
        <description>Chicago #1</description>
        <id>us-chi1</id>
    </zone>
</zones>`

	zones := Zones{}
	err := xml.Unmarshal([]byte(originalXML), &zones)

	assert.Nil(t, err)
	assert.Len(t, zones.Zones, 4)

	firstZone := zones.Zones[0]
	assert.Equal(t, "Frankfurt #1", firstZone.Description)
	assert.Equal(t, "de-fra1", firstZone.Id)
}
