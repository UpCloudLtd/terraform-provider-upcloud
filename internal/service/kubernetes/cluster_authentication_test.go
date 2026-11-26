package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseClusterAuthenticationID(t *testing.T) {
	uuid := "0a1b2c3d-0000-4000-8000-000000000000"

	clusterUUID, name, diags := parseClusterAuthenticationID(uuid + "/issuer")
	assert.False(t, diags.HasError())
	assert.Equal(t, uuid, clusterUUID.String())
	assert.Equal(t, "issuer", name)

	for _, id := range []string{"", uuid, uuid + "/", "not-a-uuid/issuer", uuid + "/issuer/extra"} {
		_, _, diags := parseClusterAuthenticationID(id)
		assert.True(t, diags.HasError(), id)
	}
}
