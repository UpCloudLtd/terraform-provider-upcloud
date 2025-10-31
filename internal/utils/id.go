package utils

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func MarshalID(components ...string) string {
	return strings.Join(components, "/")
}

func UnmarshalID(id string, components ...*string) error {
	parts := strings.Split(id, "/")
	if len(parts) > len(components) {
		return fmt.Errorf("not enough components (%d) to unmarshal id '%s'", len(components), id)
	}
	for i, c := range parts {
		*components[i] = c
	}
	return nil
}

func UnmarshalIDDiag(id string, components ...*string) (diags diag.Diagnostics) {
	err := UnmarshalID(id, components...)
	if err != nil {
		diags.AddError(
			"Unable to unmarshal ID",
			ErrorDiagnosticDetail(err),
		)
	}
	return
}
