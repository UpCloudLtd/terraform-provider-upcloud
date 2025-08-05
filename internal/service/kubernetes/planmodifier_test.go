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
			name:            "GPU plan cannot use custom_plan block",
			planType:        "GPU-8xCPU-64GB-1xL40S",
			emptyCustomPlan: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"defining `custom_plan` block with GPU-8xCPU-64GB-1xL40S plan is not supported",
					"use `custom` plan or remove the `custom_plan` block",
				),
			},
		},
		{
			name:            "Cloud Native plan cannot use custom_plan block",
			planType:        "CLOUDNATIVE-4xCPU-8GB",
			emptyCustomPlan: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"defining `custom_plan` block with CLOUDNATIVE-4xCPU-8GB plan is not supported",
					"use `custom` plan or remove the `custom_plan` block",
				),
			},
		},
		{
			name:            "Standard plan cannot use custom_plan block",
			planType:        "standard",
			emptyCustomPlan: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"defining `custom_plan` block with standard plan is not supported",
					"use `custom` plan or remove the `custom_plan` block",
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
			name:            "No errors with non-custom plan without custom_plan block",
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

func TestValidateGPUPlan(t *testing.T) {
	tests := []struct {
		name          string
		planType      string
		emptyGPUPlan  bool
		expectedDiags diag.Diagnostics
	}{
		{
			name:          "GPU plan can use gpu_plan block",
			planType:      "GPU-8xCPU-64GB-1xL40S",
			emptyGPUPlan:  false,
			expectedDiags: diag.Diagnostics{},
		},
		{
			name:          "GPU plan can omit gpu_plan block",
			planType:      "GPU-8xCPU-64GB-1xL40S",
			emptyGPUPlan:  true,
			expectedDiags: diag.Diagnostics{},
		},
		{
			name:         "Custom plan cannot use gpu_plan block",
			planType:     "custom",
			emptyGPUPlan: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"defining `gpu_plan` block with custom plan is not supported",
					"use a GPU plan (GPU-*) or remove the `gpu_plan` block",
				),
			},
		},
		{
			name:         "Standard plan cannot use gpu_plan block",
			planType:     "standard",
			emptyGPUPlan: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"defining `gpu_plan` block with standard plan is not supported",
					"use a GPU plan (GPU-*) or remove the `gpu_plan` block",
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateGPUPlan(tt.planType, tt.emptyGPUPlan)
			require.True(t, diags.Equal(tt.expectedDiags))
		})
	}
}

func TestValidateCloudNativePlan(t *testing.T) {
	tests := []struct {
		name                 string
		planType             string
		emptyCloudNativePlan bool
		expectedDiags        diag.Diagnostics
	}{
		{
			name:                 "Cloud Native plan can use cloud_native_plan block",
			planType:             "CLOUDNATIVE-4xCPU-8GB",
			emptyCloudNativePlan: false,
			expectedDiags:        diag.Diagnostics{},
		},
		{
			name:                 "Cloud Native plan can omit cloud_native_plan block",
			planType:             "CLOUDNATIVE-4xCPU-8GB",
			emptyCloudNativePlan: true,
			expectedDiags:        diag.Diagnostics{},
		},
		{
			name:                 "Custom plan cannot use cloud_native_plan block",
			planType:             "custom",
			emptyCloudNativePlan: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"defining `cloud_native_plan` block with custom plan is not supported",
					"use a Cloud Native plan (CLOUDNATIVE-*) or remove the `cloud_native_plan` block",
				),
			},
		},
		{
			name:                 "GPU plan cannot use cloud_native_plan block",
			planType:             "GPU-8xCPU-64GB-1xL40S",
			emptyCloudNativePlan: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"defining `cloud_native_plan` block with GPU-8xCPU-64GB-1xL40S plan is not supported",
					"use a Cloud Native plan (CLOUDNATIVE-*) or remove the `cloud_native_plan` block",
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateCloudNativePlan(tt.planType, tt.emptyCloudNativePlan)
			require.True(t, diags.Equal(tt.expectedDiags))
		})
	}
}
