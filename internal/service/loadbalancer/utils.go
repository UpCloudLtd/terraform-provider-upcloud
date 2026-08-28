package loadbalancer

import (
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	validNameRegexp  = regexp.MustCompile("^[a-zA-Z0-9_-]+$")
	validNameMessage = "should contain only alphanumeric characters, underscores and dashes"
	nameValidator    = stringvalidator.RegexMatches(validNameRegexp, validNameMessage)
	portValidator    = int64validator.Between(1, 65535)
)

func asBool(p *bool) basetypes.BoolValue {
	if p == nil {
		return types.BoolValue(false)
	}
	return types.BoolValue(*p)
}

func labelsMapToV9Slice(m map[string]string) []v9.LoadBalancerLabelCreate {
	return utils.LabelsMapToSliceFn(m, func(k string, v string) v9.LoadBalancerLabelCreate {
		return v9.LoadBalancerLabelCreate{
			Key:   k,
			Value: &v,
		}
	})
}

func labelsV9SliceToMap(labels []v9.LoadBalancerLabelResponse) map[string]string {
	return utils.LabelsSliceToMapFn(labels, func(label v9.LoadBalancerLabelResponse) (string, string) {
		return label.Key, utils.ValueOrEmpty(label.Value)
	})
}
