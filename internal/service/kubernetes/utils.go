package kubernetes

import (
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

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
