package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	// Plan prefixes that support custom storage configuration
	customPlanType        = "custom"
	gpuPlanPrefix         = "GPU-"
	cloudNativePlanPrefix = "CLOUDNATIVE-"
)

func getCustomPlanPlanModifier() planmodifier.List {
	return &customPlanPlanModifier{}
}

type customPlanPlanModifier struct{}

func (d *customPlanPlanModifier) Description(_ context.Context) string {
	return "Ensures that custom_plan block is set if plan field's value supports custom storage configuration (custom, GPU-, CLOUDNATIVE-) and vice versa."
}

func (d *customPlanPlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d *customPlanPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	var plan types.String
	diags := req.Plan.GetAttribute(ctx, path.Root("plan"), &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = validateCustomPlan(plan.ValueString(), req.ConfigValue.IsNull())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional validation for field restrictions based on plan type
	if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
		diags = validateCustomPlanFields(ctx, plan.ValueString(), req.ConfigValue)
		resp.Diagnostics.Append(diags...)
	}
}

// supportsCustomStorageConfig checks if a plan type supports custom storage configuration
func supportsCustomStorageConfig(plan string) bool {
	return plan == customPlanType ||
		strings.HasPrefix(plan, gpuPlanPrefix) ||
		strings.HasPrefix(plan, cloudNativePlanPrefix)
}

func validateCustomPlan(plan string, emptyCustomPlan bool) diag.Diagnostics {
	var diags diag.Diagnostics

	supportsCustomStorage := supportsCustomStorageConfig(plan)

	// Only custom plans require the custom_plan block, GPU and CloudNative plans have it optional
	if plan == customPlanType && emptyCustomPlan {
		diags.AddError(
			"`custom_plan` field is required when using custom server plan for the node group",
			"add `custom_plan` block or change to a non-custom plan",
		)

		return diags
	}

	if !supportsCustomStorage && !emptyCustomPlan {
		diags.AddError(
			fmt.Sprintf("defining `custom_plan` block with %s plan is not supported", plan),
			"use a plan that supports custom storage configuration (custom, GPU-, CLOUDNATIVE-) or remove the `custom_plan` block",
		)

		return diags
	}

	return diags
}

// validateCustomPlanFields validates that only appropriate fields are set for each plan type
func validateCustomPlanFields(_ context.Context, plan string, customPlanValue types.List) diag.Diagnostics {
	var diags diag.Diagnostics

	if customPlanValue.IsNull() || customPlanValue.IsUnknown() {
		return diags
	}

	elements := customPlanValue.Elements()
	if len(elements) == 0 {
		return diags
	}

	// Check if the custom plan object has the required/forbidden fields
	if customPlanObj, ok := elements[0].(types.Object); ok {
		attrs := customPlanObj.Attributes()

		if plan == customPlanType {
			// For custom plans, cores, memory, and storage_size are required
			if coresAttr, exists := attrs["cores"]; !exists || coresAttr.IsNull() || coresAttr.IsUnknown() {
				diags.AddError(
					"cores is required for custom plans",
					"The custom plan requires cores to be specified in the custom_plan block.",
				)
			}

			if memoryAttr, exists := attrs["memory"]; !exists || memoryAttr.IsNull() || memoryAttr.IsUnknown() {
				diags.AddError(
					"memory is required for custom plans",
					"The custom plan requires memory to be specified in the custom_plan block.",
				)
			}

			if storageSizeAttr, exists := attrs["storage_size"]; !exists || storageSizeAttr.IsNull() || storageSizeAttr.IsUnknown() {
				diags.AddError(
					"storage_size is required for custom plans",
					"The custom plan requires storage_size to be specified in the custom_plan block.",
				)
			}
		} else if strings.HasPrefix(plan, gpuPlanPrefix) || strings.HasPrefix(plan, cloudNativePlanPrefix) {
			// For GPU and Cloud Native plans, cores and memory should not be configurable
			planType := "GPU"
			if strings.HasPrefix(plan, cloudNativePlanPrefix) {
				planType = "Cloud Native"
			}

			if coresAttr, exists := attrs["cores"]; exists && !coresAttr.IsNull() && !coresAttr.IsUnknown() {
				diags.AddError(
					fmt.Sprintf("cores cannot be configured for %s plans", planType),
					fmt.Sprintf("The %s plan '%s' has fixed CPU cores. Remove the cores field from custom_plan block.", planType, plan),
				)
			}

			if memoryAttr, exists := attrs["memory"]; exists && !memoryAttr.IsNull() && !memoryAttr.IsUnknown() {
				diags.AddError(
					fmt.Sprintf("memory cannot be configured for %s plans", planType),
					fmt.Sprintf("The %s plan '%s' has fixed memory allocation. Remove the memory field from custom_plan block.", planType, plan),
				)
			}
		}
	}

	return diags
}
