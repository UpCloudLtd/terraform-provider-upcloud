package objectstorage

import (
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func createKeyValidationFunc(attrName objectStorageKeyType, minLength, maxLength int) schema.SchemaValidateDiagFunc {
	const (
		objectStorageKeyTypeAccess objectStorageKeyType = "access_key"
		objectStorageKeyTypeSecret objectStorageKeyType = "secret_key"
	)

	return func(val interface{}, path cty.Path) diag.Diagnostics {
		key, ok := val.(string)

		if !ok {
			return diag.Errorf("expected type of %v to be string", val)
		}

		// For access and secret keys empty string means that they should be taken from env vars
		if key == "" {
			var envVarPrefix string

			switch attrName {
			case objectStorageKeyTypeAccess:
				envVarPrefix = accessKeyEnvVarPrefix
			case objectStorageKeyTypeSecret:
				envVarPrefix = secretKeyEnvVarPrefix
			default:
				return diag.Errorf("unknown attribute name for creating object storage keys validation function: %s; this is a provider error", attrName)
			}

			if !utils.EnvKeyExists(envVarPrefix) {
				return diag.Errorf("%s set to empty string, but no environment variables for it found (%s{NAME})", attrName, envVarPrefix)
			}

			return diag.Diagnostics{}
		}

		length := len(key)

		if length < minLength {
			return diag.Errorf("%s too short; minimum length is %d, got %d", attrName, minLength, length)
		}

		if length > maxLength {
			return diag.Errorf("%s too long; max length is %d, got %d", attrName, maxLength, length)
		}

		return diag.Diagnostics{}
	}
}
