package kubernetes

import (
	"fmt"
	"strings"
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
