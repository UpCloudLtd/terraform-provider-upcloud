package utils

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var labelKeyRegExp = regexp.MustCompile("^([a-zA-Z0-9])+([a-zA-Z0-9_-])*$")

func labelsDescription(resource string) string {
	return fmt.Sprintf("User defined key-value pairs to classify the %s.", resource)
}

func systemLabelsDescription(resource string) string {
	return fmt.Sprintf("System defined key-value pairs to classify the %s. The keys of system defined labels are prefixed with underscore and can not be modified by the user.", resource)
}

var _ planmodifier.Map = unconfiguredAsEmpty{}

type unconfiguredAsEmpty struct{}

func (lm unconfiguredAsEmpty) Description(_ context.Context) string {
	return "use empty map, if config is null."
}

func (lm unconfiguredAsEmpty) MarkdownDescription(ctx context.Context) string {
	return lm.Description(ctx)
}

func (lm unconfiguredAsEmpty) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if req.ConfigValue.IsNull() {
		labels := make(map[string]string)
		value, diags := types.MapValueFrom(ctx, types.StringType, &labels)

		resp.PlanValue = value
		resp.Diagnostics = diags
	}
}

func ReadOnlyLabelsAttribute(resource string) schema.Attribute {
	description := labelsDescription(resource)
	return &schema.MapAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: description,
	}
}

func LabelsAttribute(resource string, additionalPlanModifiers ...planmodifier.Map) schema.Attribute {
	description := labelsDescription(resource)
	planModifiers := []planmodifier.Map{
		unconfiguredAsEmpty{},
		mapplanmodifier.UseStateForUnknown(),
	}
	planModifiers = append(planModifiers, additionalPlanModifiers...)

	return &schema.MapAttribute{
		ElementType:   types.StringType,
		Computed:      true,
		Optional:      true,
		Description:   description,
		PlanModifiers: planModifiers,
		Validators: []validator.Map{
			mapvalidator.KeysAre(stringvalidator.LengthBetween(2, 32), stringvalidator.RegexMatches(labelKeyRegExp, "")),
			mapvalidator.ValueStringsAre(stringvalidator.LengthBetween(0, 255)),
		},
	}
}

func SystemLabelsAttribute(resource string) schema.Attribute {
	description := systemLabelsDescription(resource)
	return &schema.MapAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: description,
		PlanModifiers: []planmodifier.Map{
			mapplanmodifier.UseStateForUnknown(),
		},
	}
}

func LabelsSchema(resource string) *sdkv2_schema.Schema {
	description := labelsDescription(resource)
	return &sdkv2_schema.Schema{
		Description: description,
		Type:        sdkv2_schema.TypeMap,
		Elem: &sdkv2_schema.Schema{
			Type: sdkv2_schema.TypeString,
		},
		Optional:         true,
		ValidateDiagFunc: ValidateLabelsDiagFunc,
	}
}

func LabelsMapToSlice[T any](m map[string]T) []upcloud.Label {
	var labels []upcloud.Label

	for k, v := range m {
		var value any = v
		labels = append(labels, upcloud.Label{
			Key:   k,
			Value: value.(string),
		})
	}

	return labels
}

type labelType string

var (
	labelTypeSystem labelType = "system"
	labelTypeUser   labelType = "user"
)

func labelsSliceToMap(s []upcloud.Label, t labelType) map[string]string {
	labels := make(map[string]string)

	for _, l := range s {
		if t == labelTypeSystem && strings.HasPrefix(l.Key, "_") {
			labels[l.Key] = l.Value
		}
		if t == labelTypeUser && !strings.HasPrefix(l.Key, "_") {
			labels[l.Key] = l.Value
		}
	}

	return labels
}

func LabelsSliceToMap(s []upcloud.Label) map[string]string {
	return labelsSliceToMap(s, labelTypeUser)
}

func LabelsSliceToSystemLabelsMap(s []upcloud.Label) map[string]string {
	return labelsSliceToMap(s, labelTypeSystem)
}

var ValidateLabelsDiagFunc = validation.AllDiag(
	validation.MapKeyLenBetween(2, 32),
	validation.MapKeyMatch(labelKeyRegExp, ""),
	validation.MapValueLenBetween(0, 255),
)
