package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func validatePlan(ctx context.Context, service *service.Service, plan types.String) (diags diag.Diagnostics) {
	if plan.IsNull() {
		return
	}

	plans, err := service.GetPlans(ctx)
	if err != nil {
		diags.AddError(
			"Unable to fetch available plans",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	availablePlans := make([]string, 0)
	for _, p := range plans.Plans {
		if p.Name == plan.ValueString() {
			return
		}
		availablePlans = append(availablePlans, p.Name)
	}

	diags.AddAttributeError(
		path.Root("plan"),
		"Invalid plan",
		fmt.Sprintf("expected plan to be one of [%s], got %s", strings.Join(availablePlans, ", "), plan.ValueString()),
	)
	return
}

func validateZone(ctx context.Context, service *service.Service, zone types.String) (diags diag.Diagnostics) {
	zones, err := service.GetZones(ctx)
	if err != nil {
		diags.AddError(
			"Unable to fetch available plans",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	availableZones := make([]string, 0)
	for _, z := range zones.Zones {
		if z.ID == zone.ValueString() {
			return nil
		}
		availableZones = append(availableZones, z.ID)
	}
	diags.AddAttributeError(
		path.Root("zone"),
		"Invalid zone",
		fmt.Sprintf("expected zone to be one of [%s], got %s", strings.Join(availableZones, ", "), zone.ValueString()),
	)
	return
}

type noDuplicateTagsValidator struct{}

var _ validator.Set = noDuplicateTagsValidator{}

// Description describes the validation.
func (v noDuplicateTagsValidator) Description(_ context.Context) string {
	return "must not contain case-insensitive duplicates"
}

func getTags(ctx context.Context, value basetypes.SetValue) (tags []string, diags diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		tags = nil
		return
	}
	diags.Append(value.ElementsAs(ctx, &tags, false)...)
	return
}

// MarkdownDescription describes the validation in Markdown.
func (v noDuplicateTagsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v noDuplicateTagsValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	tags, diags := getTags(ctx, req.ConfigValue)
	resp.Diagnostics.Append(diags...)

	tagsMap := make(map[string]string)
	var duplicates []string

	for _, tag := range tags {
		if duplicate, ok := tagsMap[strings.ToLower(tag)]; ok {
			duplicates = append(duplicates, fmt.Sprintf("%s = %s", duplicate, tag))
		}
		tagsMap[strings.ToLower(tag)] = tag
	}

	if len(duplicates) != 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("tags"),
			"Invalid tags",
			fmt.Sprintf("tags can not contain case-insensitive duplicates (%s)", strings.Join(duplicates, ", ")),
		)
	}
}

func validateTagsChangeRequiresMainAccount(ctx context.Context, service *service.Service, oldTagsValue, newTagsValue types.Set) (diags diag.Diagnostics) {
	oldTags, d := getTags(ctx, oldTagsValue)
	diags.Append(d...)
	newTags, diags := getTags(ctx, newTagsValue)
	diags.Append(d...)

	if tagsHasChange(oldTags, newTags) {
		isSubaccount, err := isProviderAccountSubaccount(ctx, service)
		if err != nil {
			diags.AddError(
				"Unable to determine account details",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
		if isSubaccount {
			diags.AddAttributeError(
				path.Root("tags"),
				"Unable to create or modify tags",
				fmt.Sprintf("Creating and modifying tags is allowed only for main account. Subaccounts have access only to listing tags and tagged servers they are granted access to (tags change: %v -> %v)", oldTags, newTags),
			)
		}
	}
	return
}
