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

func getGPUPlanPlanModifier() planmodifier.List {
	return &gpuPlanPlanModifier{}
}

func getCloudNativePlanPlanModifier() planmodifier.List {
	return &cloudNativePlanPlanModifier{}
}

type customPlanPlanModifier struct{}

func (d *customPlanPlanModifier) Description(_ context.Context) string {
	return "Ensures that custom_plan block is set if plan field's value is custom and vice versa."
}

type gpuPlanPlanModifier struct{}

func (d *gpuPlanPlanModifier) Description(_ context.Context) string {
	return "Ensures that gpu_plan block is only used with GPU plans."
}

type cloudNativePlanPlanModifier struct{}

func (d *cloudNativePlanPlanModifier) Description(_ context.Context) string {
	return "Ensures that cloud_native_plan block is only used with Cloud Native plans."
}

func (d *customPlanPlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d *gpuPlanPlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d *cloudNativePlanPlanModifier) MarkdownDescription(ctx context.Context) string {
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
}

func (d *gpuPlanPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	var plan types.String
	diags := req.Plan.GetAttribute(ctx, path.Root("plan"), &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = validateGPUPlan(plan.ValueString(), req.ConfigValue.IsNull())
	resp.Diagnostics.Append(diags...)
}

func (d *cloudNativePlanPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	var plan types.String
	diags := req.Plan.GetAttribute(ctx, path.Root("plan"), &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = validateCloudNativePlan(plan.ValueString(), req.ConfigValue.IsNull())
	resp.Diagnostics.Append(diags...)
}

func validateCustomPlan(plan string, emptyCustomPlan bool) diag.Diagnostics {
	var diags diag.Diagnostics

	// Only custom plans require and allow the custom_plan block
	if plan == customPlanType && emptyCustomPlan {
		diags.AddError(
			"`custom_plan` field is required when using custom server plan for the node group",
			"add `custom_plan` block or change to a non-custom plan",
		)
		return diags
	}

	if plan != customPlanType && !emptyCustomPlan {
		diags.AddError(
			fmt.Sprintf("defining `custom_plan` block with %s plan is not supported", plan),
			"use `custom` plan or remove the `custom_plan` block",
		)
		return diags
	}

	return diags
}

func validateGPUPlan(plan string, emptyGPUPlan bool) diag.Diagnostics {
	var diags diag.Diagnostics

	// GPU plan block is optional for GPU plans, not allowed for others
	if !strings.HasPrefix(plan, gpuPlanPrefix) && !emptyGPUPlan {
		diags.AddError(
			fmt.Sprintf("defining `gpu_plan` block with %s plan is not supported", plan),
			"use a GPU plan (GPU-*) or remove the `gpu_plan` block",
		)
		return diags
	}

	return diags
}

func validateCloudNativePlan(plan string, emptyCloudNativePlan bool) diag.Diagnostics {
	var diags diag.Diagnostics

	// Cloud Native plan block is optional for Cloud Native plans, not allowed for others
	if !strings.HasPrefix(plan, cloudNativePlanPrefix) && !emptyCloudNativePlan {
		diags.AddError(
			fmt.Sprintf("defining `cloud_native_plan` block with %s plan is not supported", plan),
			"use a Cloud Native plan (CLOUDNATIVE-*) or remove the `cloud_native_plan` block",
		)
		return diags
	}

	return diags
}
