package upcloud

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUnmarshalTag tests that Tag structs are unmarshaled correctly
func TestUnmarshalTag(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="utf-8"?>
<tags>
    <tag>
        <description></description>
        <name>tag1</name>
        <servers>
            <server>002f426c-a194-4624-b1fc-7e104e71fe4a</server>
        </servers>
    </tag>
    <tag>
        <description></description>
        <name>tag2</name>
        <servers>
            <server>002f426c-a194-4624-b1fc-7e104e71fe4a</server>
        </servers>
    </tag>
    <tag>
        <description></description>
        <name>tag3</name>
        <servers>
            <server>002f426c-a194-4624-b1fc-7e104e71fe4a</server>
        </servers>
    </tag>
</tags>`

	tags := Tags{}
	err := xml.Unmarshal([]byte(originalXML), &tags)

	assert.Nil(t, err)
	assert.Len(t, tags.Tags, 3)

	firstTag := tags.Tags[0]
	assert.Equal(t, "", firstTag.Description)
	assert.Equal(t, "tag1", firstTag.Name)
	assert.Len(t, firstTag.Servers, 1)
	assert.Equal(t, "002f426c-a194-4624-b1fc-7e104e71fe4a", firstTag.Servers[0])
}
