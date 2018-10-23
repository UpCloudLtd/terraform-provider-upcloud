package upcloud

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUnmarshalAccount tests that Account objects unmarshal correctly
func TestUnmarshalAccount(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="utf-8"?>
<account>
    <credits>22465.536</credits>
    <username>foobar</username>
</account>`

	account := Account{}
	err := xml.Unmarshal([]byte(originalXML), &account)
	assert.Nil(t, err)
	assert.Equal(t, 22465.536, account.Credits)
	assert.Equal(t, "foobar", account.UserName)
}
