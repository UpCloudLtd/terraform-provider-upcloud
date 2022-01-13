package upcloudschema

import (
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTerraformSchemaFromJSONSchema(t *testing.T) {
	for _, serviceType := range []upcloud.ManagedDatabaseServiceType{
		upcloud.ManagedDatabaseServiceTypePostgreSQL,
		upcloud.ManagedDatabaseServiceTypeMySQL,
	} {
		t.Run(string(serviceType), func(t *testing.T) {
			jsonSchema := ManagedDatabaseServicePropertiesSchema(serviceType)
			tfSchema := GenerateTerraformSchemaFromJSONSchema(jsonSchema, func(keyPath []string, proposedSchema *schema.Schema, source map[string]interface{}) {
				assert.NotEmpty(t, keyPath)
				assert.NotNil(t, proposedSchema)
				assert.NotNil(t, source)
			})
			if !assert.NotEmpty(t, tfSchema) {
				return
			}
		})
	}
}
