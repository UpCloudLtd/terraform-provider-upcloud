package loadbalancer

import (
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestValidateNetworks(t *testing.T) {
	tests := []struct {
		name          string
		networkModels []loadbalancerNetworkModel
		expectedDiags diag.Diagnostics
	}{
		{
			name: "Private network ID required",
			networkModels: []loadbalancerNetworkModel{
				{
					Type: types.StringValue(string(upcloud.LoadBalancerNetworkTypePrivate)),
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("load balancer's private network ID is required", "#0"),
			},
		},
		{
			name: "Public network ID not supported",
			networkModels: []loadbalancerNetworkModel{
				{
					Network: types.StringValue("not-an-empty-string"),
					Type:    types.StringValue(string(upcloud.LoadBalancerNetworkTypePublic)),
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("setting load balancer's public network ID is not supported", "#0"),
			},
		},
		{
			name: "No errors with valid network models",
			networkModels: []loadbalancerNetworkModel{
				{
					Network: types.StringValue("not-an-empty-string"),
					Type:    types.StringValue(string(upcloud.LoadBalancerNetworkTypePrivate)),
				},
				{
					Type: types.StringValue(string(upcloud.LoadBalancerNetworkTypePublic)),
				},
			},
			expectedDiags: diag.Diagnostics{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateNetworks(tt.networkModels)
			require.True(t, diags.Equal(tt.expectedDiags))
		})
	}
}
