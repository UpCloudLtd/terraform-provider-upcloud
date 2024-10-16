package kubernetes

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

func TestValidateCustomPlan(t *testing.T) {
	tests := []struct {
		name            string
		planType        string
		emptyCustomPlan bool
		expectedDiags   diag.Diagnostics
	}{
		{
			name:            "Custom plan required",
			planType:        "custom",
			emptyCustomPlan: true,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"`custom_plan` field is required when using custom server plan for the node group",
					"add `custom_plan` block or change to a non-custom plan",
				),
			},
		},
		{
			name:            "Custom plan not allowed",
			planType:        "standard",
			emptyCustomPlan: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"defining `custom_plan` block with standard plan is not supported",
					"use custom as the `plan` value or change to a non-custom plan",
				),
			},
		},
		{
			name:            "No errors with custom plan",
			planType:        "custom",
			emptyCustomPlan: false,
			expectedDiags:   diag.Diagnostics{},
		},
		{
			name:            "No errors with non-custom plan",
			planType:        "standard",
			emptyCustomPlan: true,
			expectedDiags:   diag.Diagnostics{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateCustomPlan(tt.planType, tt.emptyCustomPlan)
			require.True(t, diags.Equal(tt.expectedDiags))
		})
	}
}
