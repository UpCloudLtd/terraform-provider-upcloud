package kubernetes

import (
	"encoding/base64"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func setB64Decoded(d *schema.ResourceData, field string, value string) error {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return err
	}

	return d.Set(field, string(decoded))
}

func storageEncryptionSchema(description string, computed bool) *schema.Schema {
	return &schema.Schema{
		Description: description,
		Type:        schema.TypeString,
		Computed:    computed,
		Optional:    true,
		ForceNew:    true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
			string(upcloud.StorageEncryptionDataAtReset),
		}, false)),
	}
}
