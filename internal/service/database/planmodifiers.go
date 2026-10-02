package database

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type rfc3339Validator struct{}

var _ validator.String = rfc3339Validator{}

func (v rfc3339Validator) Description(_ context.Context) string {
	return "must be a valid RFC 3339 timestamp"
}

func (v rfc3339Validator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v rfc3339Validator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if _, err := time.Parse(time.RFC3339, req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid timestamp",
			"Value must be a valid RFC 3339 timestamp, e.g. `2026-01-02T15:04:05Z`: "+err.Error(),
		)
	}
}

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
