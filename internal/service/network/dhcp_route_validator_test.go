package network

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestDHCPRouteValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		hasError bool
	}{
		{
			input:    "192.168.0.0/24-nexthop=10.0.1.100",
			hasError: false,
		},
		{
			input:    "192.168.0.0/24-nexthop=10.0.1.0/24",
			hasError: true,
		},
		{
			input:    "192.168.0.0/24",
			hasError: false,
		},
		{
			input:    "192.168.0.0",
			hasError: true,
		},
		{
			input:    "Cow says moo! 🐄",
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			v := dhcpRouteValidator{}
			req := validator.StringRequest{
				ConfigValue: types.StringValue(test.input),
			}
			resp := validator.StringResponse{}
			v.ValidateString(context.TODO(), req, &resp)
			errorDetails := []string{}
			for _, diag := range resp.Diagnostics.Errors() {
				errorDetails = append(errorDetails, diag.Detail())
			}
			assert.Equal(t, test.hasError, resp.Diagnostics.HasError(), "Errors: %s", strings.Join(errorDetails, ", "))
		})
	}
}
