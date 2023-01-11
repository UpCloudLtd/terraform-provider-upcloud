package loadbalancer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func marshalID(components ...string) string {
	return strings.Join(components, "/")
}

func unmarshalID(id string, components ...*string) error {
	parts := strings.Split(id, "/")
	if len(parts) > len(components) {
		return fmt.Errorf("not enough components (%d) to unmarshal id '%s'", len(components), id)
	}
	for i, c := range parts {
		*components[i] = c
	}
	return nil
}

var validateNameDiagFunc = validation.ToDiagFunc(validation.StringMatch(
	regexp.MustCompile("^[a-zA-Z0-9_-]+$"),
	"should contain only alphanumeric characters, underscores and dashes",
))
