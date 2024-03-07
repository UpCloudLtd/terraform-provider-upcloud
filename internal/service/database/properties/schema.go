package properties

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ensureDot(text string) string {
	if text == "" {
		return text
	}
	if strings.HasSuffix(text, ".") {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(text) + "."
}

func getDescription(prop upcloud.ManagedDatabaseServiceProperty) string {
	return strings.TrimSpace(ensureDot(prop.Title) + " " + ensureDot(prop.Description))
}

func getType(raw interface{}) string {
	if types, ok := raw.([]interface{}); ok {
		for _, t := range types {
			if t.(string) != "null" {
				return t.(string)
			}
		}
	}
	return raw.(string)
}

func diffSuppressCreateOnlyProperty(_, _, _ string, d *schema.ResourceData) bool {
	return d.Id() != ""
}

func stringSlice(val interface{}) ([]string, bool) {
	if strs, ok := val.([]string); ok {
		return strs, true
	}
	if ifaces, ok := val.([]interface{}); ok {
		strs := make([]string, 0)
		for _, iface := range ifaces {
			str, ok := iface.(string)
			if !ok {
				return nil, false
			}
			strs = append(strs, str)
		}
		return strs, true
	}
	return nil, false
}

func getSchema(prop upcloud.ManagedDatabaseServiceProperty) (*schema.Schema, error) {
	s := schema.Schema{
		Description: getDescription(prop),
		Optional:    true,
		Computed:    true,
	}

	if prop.CreateOnly {
		s.ForceNew = true
		s.DiffSuppressFunc = diffSuppressCreateOnlyProperty
	}

	validations := []schema.SchemaValidateFunc{} //nolint:staticcheck // Most of the validators still use the deprecated function signature.

	switch getType(prop.Type) {
	case "string":
		s.Type = schema.TypeString

		if prop.MaxLength != 0 {
			validations = append(validations, validation.StringLenBetween(prop.MinLength, prop.MaxLength))
		}
		if enum, ok := stringSlice(prop.Enum); ok {
			validations = append(validations, validation.StringInSlice(enum, false))
		}
		if prop.Pattern != "" {
			if re, err := regexp.Compile(prop.Pattern); err == nil {
				validations = append(validations, validation.StringMatch(re, fmt.Sprintf(`Must match "%s" pattern.`, prop.Pattern)))
			}
		}
	case "integer":
		s.Type = schema.TypeInt

		if prop.Minimum != nil {
			validations = append(validations, validation.IntAtLeast(int(*prop.Minimum)))
		}
		if prop.Maximum != nil {
			validations = append(validations, validation.IntAtMost(int(*prop.Maximum)))
		}
	case "number":
		s.Type = schema.TypeFloat

		if prop.Minimum != nil {
			validations = append(validations, validation.FloatAtLeast(*prop.Minimum))
		}
		if prop.Maximum != nil {
			validations = append(validations, validation.FloatAtMost(*prop.Maximum))
		}
	case "boolean":
		s.Type = schema.TypeBool
	case "array":
		s.Type = schema.TypeList
		s.Elem = &schema.Schema{
			Type: schema.TypeString,
		}
	case "object":
		nested, err := getSchemaMap(prop.Properties)
		if err != nil {
			return nil, err
		}

		s.Type = schema.TypeList
		s.MaxItems = 1
		s.Elem = &schema.Resource{Schema: nested}
	default:
		return nil, fmt.Errorf(`unknown property value type "%s"`, prop.Type)
	}

	if len(validations) > 0 {
		s.ValidateDiagFunc = validation.ToDiagFunc(
			validation.All(validations...),
		)
	}

	return &s, nil
}

func getSchemaMap(props map[string]upcloud.ManagedDatabaseServiceProperty) (map[string]*schema.Schema, error) {
	sMap := make(map[string]*schema.Schema)
	for key, prop := range props {
		s, err := getSchema(prop)
		if err != nil {
			return nil, err
		}
		sMap[key] = s
	}
	return sMap, nil
}

func panicMessage(dbType upcloud.ManagedDatabaseServiceType, step string) string {
	return fmt.Sprintf("Could not generate %s properties %s. This is a bug in the provider. Please create an issue in https://github.com/UpCloudLtd/terraform-provider-upcloud/issues.", dbType, step)
}

func GetSchemaMap(dbType upcloud.ManagedDatabaseServiceType) map[string]*schema.Schema {
	p, err := getPropertiesMap(dbType)
	if err != nil {
		panic(panicMessage(dbType, "map"))
	}

	s, err := getSchemaMap(p)
	if err != nil {
		panic(panicMessage(dbType, "schema"))
	}
	return s
}
