package upcloudschema

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// FnGenerateTerraformSchemaOverride can be used to override the schema generation of GenerateTerraformSchemaFromJSONSchema.
// As a schema can contain multiple nested sub-schemas, passing a function with this signature allows one to inspect the
// current key and the proposed schema. The function can then change the proposedSchema as they see fit.
type FnGenerateTerraformSchemaOverride func(keyPath []string, proposedSchema *schema.Schema, source map[string]interface{})

// GenerateTerraformSchemaFromJSONSchema generates a Terraform schema from a subset of JSON schema
// The root json schema must be of type "object" and this function expects you to pass the "properties" field of the
// schema object.
func GenerateTerraformSchemaFromJSONSchema(objectProperties map[string]interface{}, override FnGenerateTerraformSchemaOverride) map[string]*schema.Schema {
	return generateTerraformSchemaFromJSONSchema([]string{}, objectProperties, override)
}

func generateTerraformSchemaFromJSONSchema(keyPath []string, objectProperties map[string]interface{}, override FnGenerateTerraformSchemaOverride) map[string]*schema.Schema {
	keyReplacer := strings.NewReplacer(".", "_")
	r := make(map[string]*schema.Schema)
	for k, v := range objectProperties {
		k = keyReplacer.Replace(k)
		newKeyPath := append([]string{}, keyPath...)
		newKeyPath = append(newKeyPath, k)
		r[k] = terraformSchemaForSingle(newKeyPath, v.(map[string]interface{}), override)
	}
	return r
}

func terraformSchemaForSingle(keyPath []string, jsonSchema map[string]interface{}, override FnGenerateTerraformSchemaOverride) *schema.Schema {
	typeForJSONSchemaType := func(jsonSchemaType interface{}) (r string) {
		switch v := jsonSchemaType.(type) {
		case string:
			r = v
		case []interface{}:
			for _, jsonType := range v {
				if _, ok := jsonType.(string); !ok {
					panic(fmt.Sprintf("invalid json type value %T (%v)", jsonType, jsonSchema))
				}
				if jsonType == "null" {
					continue
				}
				r = jsonType.(string)
				break
			}
		default:
			panic(fmt.Sprintf("invalid json type value %T (%v) (key %s)", v, v, strings.Join(keyPath, ".")))
		}
		return r
	}

	r := &schema.Schema{
		Description: jsonSchema["title"].(string),
	}

	valueType := typeForJSONSchemaType(jsonSchema["type"])
	switch valueType {
	case "number":
		r.Type = schema.TypeFloat
	case "integer":
		r.Type = schema.TypeInt
	case "string":
		r.Type = schema.TypeString
	case "boolean":
		r.Type = schema.TypeBool
	case "object":
		r.Type = schema.TypeList
		r.MaxItems = 1
		r.Elem = &schema.Resource{Schema: generateTerraformSchemaFromJSONSchema(keyPath, jsonSchema["properties"].(map[string]interface{}), override)}
	case "array":
		r.Type = schema.TypeList
		if v, ok := jsonSchema["maxItems"]; ok {
			r.MaxItems = int(v.(float64))
		}
		itemValueSchema := jsonSchema["items"].(map[string]interface{})
		if itemValueSchema["type"] == "object" {
			r.Elem = &schema.Resource{Schema: generateTerraformSchemaFromJSONSchema(keyPath, itemValueSchema["properties"].(map[string]interface{}), override)}
		} else {
			r.Elem = terraformSchemaForSingle(nil, itemValueSchema, nil)
		}
	default:
		panic(fmt.Sprintf("non-supported json schema type %v (key %s)", valueType, strings.Join(keyPath, ".")))
	}

	if override != nil {
		override(keyPath, r, jsonSchema)
	}
	return r
}
