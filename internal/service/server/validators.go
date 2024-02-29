package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/validator"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func validateHostnameDiagFunc(min, max int) schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		val, ok := v.(string)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Bad type",
				Detail:        "expected type to be string",
				AttributePath: path,
			})
			return diags
		}

		if len(val) < min || len(val) > max {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Hostname length validation failed",
				Detail:        fmt.Sprintf("expected hostname length to be in the range (%d - %d), got %d", min, max, len(val)),
				AttributePath: path,
			})
			return diags
		}

		if err := validator.ValidateDomainName(val); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Hostname validation failed",
				Detail:        err.Error(),
				AttributePath: path,
			})
		}

		return diags
	}
}

func validatePlan(ctx context.Context, service *service.Service, plan string) error {
	if plan == "" {
		return nil
	}
	plans, err := service.GetPlans(ctx)
	if err != nil {
		return err
	}
	availablePlans := make([]string, 0)
	for _, p := range plans.Plans {
		if p.Name == plan {
			return nil
		}
		availablePlans = append(availablePlans, p.Name)
	}
	return fmt.Errorf("expected plan to be one of [%s], got %s", strings.Join(availablePlans, ", "), plan)
}

func validateZone(ctx context.Context, service *service.Service, zone string) error {
	zones, err := service.GetZones(ctx)
	if err != nil {
		return err
	}
	availableZones := make([]string, 0)
	for _, z := range zones.Zones {
		if z.ID == zone {
			return nil
		}
		availableZones = append(availableZones, z.ID)
	}
	return fmt.Errorf("expected zone to be one of [%s], got %s", strings.Join(availableZones, ", "), zone)
}

func validateTagsChange(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	oldTags, newTags := d.GetChange("tags")
	if tagsHasChange(oldTags, newTags) {
		client := meta.(*service.Service)

		if isSubaccount, err := isProviderAccountSubaccount(ctx, client); err != nil || isSubaccount {
			if err != nil {
				return err
			}
			return fmt.Errorf("creating and modifying tags is allowed only by main account. Subaccounts have access only to listing tags and tagged servers they are granted access to (tags change: %v -> %v)", oldTags, newTags)
		}
	}

	tagsMap := make(map[string]string)
	var duplicates []string

	for _, tag := range utils.ExpandStrings(d.Get("tags")) {
		if duplicate, ok := tagsMap[strings.ToLower(tag)]; ok {
			duplicates = append(duplicates, fmt.Sprintf("%s = %s", duplicate, tag))
		}
		tagsMap[strings.ToLower(tag)] = tag
	}

	if len(duplicates) != 0 {
		return fmt.Errorf("tags can not contain case-insensitive duplicates (%s)", strings.Join(duplicates, ", "))
	}

	return nil
}
