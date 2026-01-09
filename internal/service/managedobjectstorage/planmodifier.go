package managedobjectstorage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.String = useStateAfterCreateStringPlanModifier{}

type useStateAfterCreateStringPlanModifier struct{}

func UseStateAfterCreate() planmodifier.String {
	return useStateAfterCreateStringPlanModifier{}
}

func (m useStateAfterCreateStringPlanModifier) Description(_ context.Context) string {
	return "Keeps the existing state value after creation (create-only behavior)."
}

func (m useStateAfterCreateStringPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStateAfterCreateStringPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If state is unknown/null, we are in create (or import planning); allow config to set it.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	// If config is unknown, keep state.
	if req.PlanValue.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}

	// If config differs from state after creation, ignore changes by keeping state.
	// It prevents Terraform from planning an in-place update for a create-only field.
	if !req.PlanValue.Equal(req.StateValue) {
		resp.PlanValue = req.StateValue
		return
	}
}
