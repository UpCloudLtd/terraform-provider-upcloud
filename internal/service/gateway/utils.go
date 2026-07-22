package gateway

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var validateName = validation.ToDiagFunc(validation.All(
	validation.StringLenBetween(1, 64),
	validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]+$"), "must contain only alphanumeric characters, hyphens, and underscores"),
))
