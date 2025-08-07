package kubernetes

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
			name:            "GPU plan does not require custom_plan block",
			planType:        "GPU-8xCPU-64GB-1xL40S",
			emptyCustomPlan: true,
			expectedDiags:   diag.Diagnostics{},
		},
		{
			name:            "Cloud Native plan does not require custom_plan block",
			planType:        "CLOUDNATIVE-4xCPU-8GB",
			emptyCustomPlan: true,
			expectedDiags:   diag.Diagnostics{},
		},
		{
			name:            "Custom plan not allowed",
			planType:        "standard",
			emptyCustomPlan: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"defining `custom_plan` block with standard plan is not supported",
					"use a plan that supports custom storage configuration (custom, GPU-, CLOUDNATIVE-) or remove the `custom_plan` block",
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
			name:            "No errors with GPU plan",
			planType:        "GPU-8xCPU-64GB-1xL40S",
			emptyCustomPlan: false,
			expectedDiags:   diag.Diagnostics{},
		},
		{
			name:            "No errors with Cloud Native plan",
			planType:        "CLOUDNATIVE-4xCPU-8GB",
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

func TestSupportsCustomStorageConfig(t *testing.T) {
	tests := []struct {
		name     string
		planType string
		expected bool
	}{
		{
			name:     "Custom plan supports custom storage",
			planType: "custom",
			expected: true,
		},
		{
			name:     "GPU plan supports custom storage",
			planType: "GPU-8xCPU-64GB-1xL40S",
			expected: true,
		},
		{
			name:     "Cloud Native plan supports custom storage",
			planType: "CLOUDNATIVE-4xCPU-8GB",
			expected: true,
		},
		{
			name:     "Standard plan does not support custom storage",
			planType: "standard",
			expected: false,
		},
		{
			name:     "Development plan does not support custom storage",
			planType: "development",
			expected: false,
		},
		{
			name:     "2xCPU-4GB plan does not support custom storage",
			planType: "2xCPU-4GB",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := supportsCustomStorageConfig(tt.planType)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateCustomPlanFields(t *testing.T) {
	tests := []struct {
		name            string
		planType        string
		customPlanAttrs map[string]attr.Value
		expectedDiags   diag.Diagnostics
	}{
		{
			name:     "Custom plan with all required fields",
			planType: "custom",
			customPlanAttrs: map[string]attr.Value{
				"cores":        types.Int64Value(4),
				"memory":       types.Int64Value(8192),
				"storage_size": types.Int64Value(100),
				"storage_tier": types.StringValue("maxiops"),
			},
			expectedDiags: diag.Diagnostics{},
		},
		{
			name:     "Custom plan missing cores",
			planType: "custom",
			customPlanAttrs: map[string]attr.Value{
				"memory":       types.Int64Value(8192),
				"storage_size": types.Int64Value(100),
				"storage_tier": types.StringValue("maxiops"),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"cores is required for custom plans",
					"The custom plan requires cores to be specified in the custom_plan block.",
				),
			},
		},
		{
			name:     "GPU plan with cores should fail",
			planType: "GPU-8xCPU-64GB-1xL40S",
			customPlanAttrs: map[string]attr.Value{
				"cores":        types.Int64Value(8),
				"storage_size": types.Int64Value(500),
				"storage_tier": types.StringValue("maxiops"),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"cores cannot be configured for GPU plans",
					"The GPU plan 'GPU-8xCPU-64GB-1xL40S' has fixed CPU cores. Remove the cores field from custom_plan block.",
				),
			},
		},
		{
			name:     "GPU plan with only storage configuration",
			planType: "GPU-8xCPU-64GB-1xL40S",
			customPlanAttrs: map[string]attr.Value{
				"storage_size": types.Int64Value(500),
				"storage_tier": types.StringValue("maxiops"),
			},
			expectedDiags: diag.Diagnostics{},
		},
		{
			name:     "Cloud Native plan with memory should fail",
			planType: "CLOUDNATIVE-4xCPU-8GB",
			customPlanAttrs: map[string]attr.Value{
				"memory":       types.Int64Value(8192),
				"storage_size": types.Int64Value(100),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"memory cannot be configured for Cloud Native plans",
					"The Cloud Native plan 'CLOUDNATIVE-4xCPU-8GB' has fixed memory allocation. Remove the memory field from custom_plan block.",
				),
			},
		},
		{
			name:     "Cloud Native plan with only storage_tier",
			planType: "CLOUDNATIVE-4xCPU-8GB",
			customPlanAttrs: map[string]attr.Value{
				"storage_tier": types.StringValue("standard"),
			},
			expectedDiags: diag.Diagnostics{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a types.Object from the attributes
			objType := types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"cores":        types.Int64Type,
					"memory":       types.Int64Type,
					"storage_size": types.Int64Type,
					"storage_tier": types.StringType,
				},
			}

			// Fill in missing attributes with null values
			allAttrs := map[string]attr.Value{
				"cores":        types.Int64Null(),
				"memory":       types.Int64Null(),
				"storage_size": types.Int64Null(),
				"storage_tier": types.StringNull(),
			}
			for k, v := range tt.customPlanAttrs {
				allAttrs[k] = v
			}

			obj, _ := types.ObjectValue(objType.AttrTypes, allAttrs)
			customPlanList, _ := types.ListValue(objType, []attr.Value{obj})

			diags := validateCustomPlanFields(context.Background(), tt.planType, customPlanList)
			require.True(t, diags.Equal(tt.expectedDiags))
		})
	}
}
