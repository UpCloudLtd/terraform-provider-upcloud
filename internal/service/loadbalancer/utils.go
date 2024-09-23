package loadbalancer

import (
	"regexp"

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
