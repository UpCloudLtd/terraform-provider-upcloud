package loadbalancer

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var validateNameDiagFunc = validation.ToDiagFunc(validation.StringMatch(
	regexp.MustCompile("^[a-zA-Z0-9_-]+$"),
	"should contain only alphanumeric characters, underscores and dashes",
))
