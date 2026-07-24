package managedobjectstorage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const errorPageStatusMatcherValidationDetail = "Provide either status_code or both status_range_start and status_range_end, but not both."

type errorPagesValidator struct{}

var _ validator.List = errorPagesValidator{}

func (v errorPagesValidator) Description(_ context.Context) string {
	return errorPageStatusMatcherValidationDetail
}

func (v errorPagesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v errorPagesValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	resp.Diagnostics.Append(validateErrorPages(ctx, req.ConfigValue, req.Path)...)
}

func validateErrorPages(ctx context.Context, list types.List, basePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if list.IsNull() || list.IsUnknown() {
		return diags
	}

	var pages []errorPageModel
	diags.Append(list.ElementsAs(ctx, &pages, false)...)
	if diags.HasError() {
		return diags
	}

	for i, page := range pages {
		if errorPageStatusMatcherIsValid(page) {
			continue
		}

		diags.AddAttributeError(
			basePath.AtListIndex(i),
			"Invalid error_pages status matcher",
			errorPageStatusMatcherValidationDetail,
		)
	}

	return diags
}

func errorPageStatusMatcherIsValid(page errorPageModel) bool {
	hasStatusCode, statusCodeUnknown := errorPageInt64IsSet(page.StatusCode)
	hasRangeStart, rangeStartUnknown := errorPageInt64IsSet(page.StatusRangeStart)
	hasRangeEnd, rangeEndUnknown := errorPageInt64IsSet(page.StatusRangeEnd)

	if statusCodeUnknown || rangeStartUnknown || rangeEndUnknown {
		return true
	}

	hasRange := hasRangeStart && hasRangeEnd

	return (hasStatusCode && !hasRangeStart && !hasRangeEnd) || (!hasStatusCode && hasRange)
}

func errorPageInt64IsSet(value types.Int64) (bool, bool) {
	if value.IsUnknown() {
		return false, true
	}

	return !value.IsNull(), false
}

func validateErrorPageStatusMatcherAtIndex(pages []errorPageModel) error {
	for i, page := range pages {
		if errorPageStatusMatcherIsValid(page) {
			continue
		}

		return fmt.Errorf("error_pages[%d]: %s", i, errorPageStatusMatcherValidationDetail)
	}

	return nil
}
