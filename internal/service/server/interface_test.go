package server

import (
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveInterfaceIPAddress(t *testing.T) {
	want := ""
	got, err := resolveInterfaceIPAddress(&upcloud.Network{
		UUID: "111-222-333",
		Name: "test net",
		IPNetworks: []upcloud.IPNetwork{
			{
				Address: "10.0.1.0/24",
				DHCP:    0,
			},
			{
				Address: "10.0.2.0/24",
				DHCP:    1,
			},
		},
	}, "10.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, want, got)

	want = "10.0.0.2"
	got, err = resolveInterfaceIPAddress(&upcloud.Network{
		UUID: "111-222-333",
		Name: "test net",
		IPNetworks: []upcloud.IPNetwork{
			{
				Address: "10.0.0.0/24",
				DHCP:    0,
			},
		},
	}, want)
	require.NoError(t, err)
	assert.Equal(t, want, got)

	want = "10.0.1.2"
	_, err = resolveInterfaceIPAddress(&upcloud.Network{
		UUID: "111-222-333",
		Name: "test net",
		IPNetworks: []upcloud.IPNetwork{
			{
				Address: "10.0.0.0/24",
				DHCP:    0,
			},
		},
	}, want)
	assert.Error(t, err)
}
