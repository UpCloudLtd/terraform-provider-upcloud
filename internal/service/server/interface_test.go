package server

import (
	"context"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func TestShouldModifyInterface_addAdditionalIPAddress(t *testing.T) {
	aaPlan := []additionalIPAddressModel{
		{
			IPAddress:         types.StringValue("10.100.10.3"),
			IPAddressFamily:   types.StringValue("IPv4"),
			IPAddressFloating: types.BoolValue(false),
		},
		{
			IPAddress:         types.StringValue("10.100.10.4"),
			IPAddressFamily:   types.StringValue("IPv4"),
			IPAddressFloating: types.BoolValue(false),
		},
	}
	aaType := types.ObjectType{
		AttrTypes: additionalIPAddressModel{}.AttributeTypes(),
	}
	aaList, d := types.SetValueFrom(context.TODO(), aaType, aaPlan)
	require.False(t, d.HasError())

	assert.True(t, shouldModifyInterface(
		networkInterfaceModel{
			Index:                 types.Int64Value(1),
			Type:                  types.StringValue("private"),
			Network:               types.StringValue("111-222-333"),
			IPAddress:             types.StringValue("10.100.10.2"),
			IPAddressFamily:       types.StringValue("IPv4"),
			AdditionalIPAddresses: aaList,
		},
		request.CreateNetworkInterfaceIPAddressSlice{
			{
				Family:  "IPv4",
				Address: "10.100.10.2",
			},
			{
				Family:  "IPv4",
				Address: "10.100.10.3",
			},
			{
				Family:  "IPv4",
				Address: "10.100.10.4",
			},
		},
		&upcloud.ServerInterface{
			Index:   1,
			Type:    upcloud.NetworkTypePrivate,
			Network: "111-222-333",
			IPAddresses: upcloud.IPAddressSlice{
				{
					Family:  "IPv4",
					Address: "10.100.10.2",
				},
				{
					Family:  "IPv4",
					Address: "10.100.10.3",
				},
			},
		},
	))
}
