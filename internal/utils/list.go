package utils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type elementsAser interface {
	ElementsAs(context.Context, interface{}, bool) diag.Diagnostics
}

func GetFirstItem[T any](ctx context.Context, list elementsAser) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics

	var items []T
	diags.Append(list.ElementsAs(ctx, &items, false)...)
	if diags.HasError() {
		return nil, diags
	}

	if len(items) < 1 {
		return nil, diags
	}

	return &items[0], diags
}
