package database

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type mustBeTrueOnCreate struct{}

// This is planmodifier instead of validator because we need access to the state.
var _ planmodifier.Bool = mustBeTrueOnCreate{}

// Description describes the validation.
func (v mustBeTrueOnCreate) Description(_ context.Context) string {
	return "must be true when creating the resource"
}

// MarkdownDescription describes the validation in Markdown.
func (v mustBeTrueOnCreate) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v mustBeTrueOnCreate) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.State.Raw.IsNull() && !req.PlanValue.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid value",
			"Attribute must be set to true when creating the resource.",
		)
	}
}
