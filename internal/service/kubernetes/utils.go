package kubernetes

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func setB64Decoded(d *schema.ResourceData, field string, value string) error {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return err
	}

	return d.Set(field, string(decoded))
}
