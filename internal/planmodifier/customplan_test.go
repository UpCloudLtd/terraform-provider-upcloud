package planmodifier

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomPlan_CustomPlanRequired(t *testing.T) {
	diags := validateCustomPlan("custom", true)
	require.True(t, diags.HasError())
	summary := diags[0].Summary()
	require.Equal(t, "`custom_plan` field is required when using custom server plan for the node group", summary)
}

func TestCustomPlan_CustomPlanNotAllowed(t *testing.T) {
	diags := validateCustomPlan("standard", false)
	require.True(t, diags.HasError())
	summary := diags[0].Summary()
	require.Equal(t, "defining `custom_plan` block with standard plan is not supported", summary)
}

func TestCustomPlan_NoErrorsWithCustomPlan(t *testing.T) {
	diags := validateCustomPlan("custom", false)
	require.False(t, diags.HasError())
}

func TestCustomPlan_NoErrorsWithNonCustomPlan(t *testing.T) {
	diags := validateCustomPlan("standard", true)
	require.False(t, diags.HasError())
}
