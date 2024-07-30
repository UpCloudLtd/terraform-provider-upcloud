package planmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func CustomPlanPlanModifier() planmodifier.List {
	return &customPlanPlanModifier{}
}

type customPlanPlanModifier struct{}

func (d *customPlanPlanModifier) Description(_ context.Context) string {
	return "Ensures that custom_plan block is set if plan field's value is custom and vice versa."
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

	if plan.ValueString() == "custom" && req.ConfigValue.IsNull() {
		resp.Diagnostics.AddError(
			"`custom_plan` field is required when using custom server plan for the node group",
			"add `custom_plan` block or change to a non-custom plan",
		)
		return
	}

	if plan.ValueString() != "custom" && !req.ConfigValue.IsNull() {
		resp.Diagnostics.AddError(
			fmt.Sprintf("defining `custom_plan` block with %s plan is not supported", plan),
			"use custom as the `plan` value or change to a non-custom plan",
		)
		return
	}
}
