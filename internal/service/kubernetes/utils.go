package kubernetes

import (
	"encoding/base64"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func setB64Decoded(d *schema.ResourceData, field string, value string) error {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return err
	}

	return d.Set(field, string(decoded))
}
