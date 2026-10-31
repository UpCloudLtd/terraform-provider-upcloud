package database

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// connectionProperties lists the properties whose value is part of the connection details
// (e.g. the URI scheme). A change to any of these may change the connection attributes.
var connectionProperties = []string{"valkey_ssl"}

type connectionDetailsUseStateForUnknown struct{}

// connectionDetailsUseStateForUnknown behaves like stringplanmodifier.UseStateForUnknown, but
// only when the update cannot change the connection details: the service stays powered as it
// is and none of the connectionProperties change. Otherwise the value is left unknown so that
// Terraform does not report an inconsistent result after apply.
var _ planmodifier.String = connectionDetailsUseStateForUnknown{}

func (m connectionDetailsUseStateForUnknown) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change unless the power state or a connection related property changes."
}

func (m connectionDetailsUseStateForUnknown) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m connectionDetailsUseStateForUnknown) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	changed, diags := connectionSettingsChanged(ctx, req.State, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || changed {
		return
	}

	resp.PlanValue = req.StateValue
}

type attributeGetter interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
}

// connectionSettingsChanged reports whether the planned update may change the connection
// details of the service. Unknown planned values are treated as changes.
func connectionSettingsChanged(ctx context.Context, state, plan attributeGetter) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	var statePowered, planPowered types.Bool
	diags.Append(state.GetAttribute(ctx, path.Root("powered"), &statePowered)...)
	diags.Append(plan.GetAttribute(ctx, path.Root("powered"), &planPowered)...)
	if diags.HasError() {
		return true, diags
	}
	if planPowered.IsUnknown() || !planPowered.Equal(statePowered) {
		return true, diags
	}

	var stateProps, planProps types.List
	diags.Append(state.GetAttribute(ctx, path.Root("properties"), &stateProps)...)
	diags.Append(plan.GetAttribute(ctx, path.Root("properties"), &planProps)...)
	if diags.HasError() {
		return true, diags
	}
	if planProps.IsUnknown() {
		return true, diags
	}
	// A null properties block in the plan does not send any property changes.
	if planProps.IsNull() {
		return false, diags
	}

	for _, key := range connectionProperties {
		stateValue := propertyValue(stateProps, key)
		planValue := propertyValue(planProps, key)
		if planValue == nil && stateValue == nil {
			continue
		}
		if planValue == nil || stateValue == nil || planValue.IsUnknown() || !planValue.Equal(stateValue) {
			return true, diags
		}
	}

	return false, diags
}

// propertyValue returns the value of the given key in the first element of the properties
// list, or nil if the list is empty or the key does not exist in the properties object.
func propertyValue(props types.List, key string) attr.Value {
	if props.IsNull() || props.IsUnknown() || len(props.Elements()) == 0 {
		return nil
	}
	obj, ok := props.Elements()[0].(types.Object)
	if !ok {
		return nil
	}
	return obj.Attributes()[key]
}
