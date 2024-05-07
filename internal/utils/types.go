package utils

import (
	"context"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func AsUpCloudBoolean(b types.Bool) upcloud.Boolean {
	if b.IsNull() || b.IsUnknown() {
		return upcloud.Empty
	}

	if b.ValueBool() {
		return upcloud.True
	}

	return upcloud.False
}

func AsBool(b upcloud.Boolean) types.Bool {
	if b.Empty() {
		return types.BoolPointerValue(nil)
	}

	return types.BoolValue(b.Bool())
}

func SetAsSliceOfStrings(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}

	elements := make([]types.String, 0, len(set.Elements()))
	strings := make([]string, 0, len(set.Elements()))
	diags := set.ElementsAs(ctx, &elements, false)

	for _, element := range elements {
		strings = append(strings, element.ValueString())
	}

	return strings, diags
}

func NilAsEmptyList[T any](l []T) []T {
	if l == nil {
		return []T{}
	}

	return l
}
