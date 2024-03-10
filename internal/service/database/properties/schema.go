package properties

import (
	"fmt"
	"math"
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

func titleSimilarToKey(key, title string) bool {
	if title == key {
		return true
	}

	// E.g. title = "pg_stat_statements.track" and key = "pg_stat_statements_track"
	if strings.ReplaceAll(title, ".", "_") == key {
		return true
	}

	// E.g. title == "timescaledb.max_background_workers" and key = "max_background_workers"
	if _, childTitle, ok := strings.Cut(title, "."); ok && childTitle == key {
		return true
	}

	return false
}

func getDescription(key string, prop upcloud.ManagedDatabaseServiceProperty) string {
	if titleSimilarToKey(key, prop.Title) {
		return ensureDot(prop.Description)
	}

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

func booleanDefault(val interface{}) (bool, bool) {
	if b, ok := val.(bool); ok {
		return b, true
	}

	return false, false
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

func getKeyDiffSuppressFunc(key string) schema.SchemaDiffSuppressFunc {
	switch key {
	case "ip_filter":
		return func(k, old, new string, d *schema.ResourceData) bool {
			return strings.TrimSuffix(old, "/32") == strings.TrimSuffix(new, "/32")
		}
	default:
		return nil
	}
}

func isSensitive(key string) bool {
	return strings.Contains(key, "password")
}

func getSchema(key string, prop upcloud.ManagedDatabaseServiceProperty) (*schema.Schema, error) {
	s := schema.Schema{
		Description: getDescription(key, prop),
		Optional:    true,
		Computed:    true,
		Sensitive:   isSensitive(key),
	}

	if prop.CreateOnly {
		s.ForceNew = true
		s.DiffSuppressFunc = diffSuppressCreateOnlyProperty
	}

	validations := []schema.SchemaValidateFunc{} //nolint:staticcheck // Most of the validators still use the deprecated function signature.

	switch getType(prop.Type) {
	case "string":
		s.Type = schema.TypeString
		var hasEnum, patternMatchesEmpty bool

		if prop.MaxLength != 0 {
			validations = append(validations, validation.StringLenBetween(prop.MinLength, prop.MaxLength))
		}
		if enum, hasEnum := stringSlice(prop.Enum); hasEnum {
			validations = append(validations, validation.StringInSlice(enum, false))
		}
		if prop.Pattern != "" {
			if re, err := regexp.Compile(prop.Pattern); err == nil {
				patternMatchesEmpty = re.Match([]byte{})
				validations = append(validations, validation.StringMatch(re, fmt.Sprintf(`Must match "%s" pattern.`, prop.Pattern)))
			}
		}

		// If empty string is valid value, `Computed = true` will block user from clearing the property value
		if prop.MinLength == 0 && !hasEnum && patternMatchesEmpty {
			s.Computed = false
		}

	case "integer":
		s.Type = schema.TypeInt

		if prop.Minimum != nil {
			validations = append(validations, validation.IntAtLeast(int(*prop.Minimum)))
		}
		if prop.Maximum != nil {
			if *prop.Maximum <= float64(math.MaxInt) {
				validations = append(validations, validation.IntAtMost(int(*prop.Maximum)))
			}
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

		if boolDefault, ok := booleanDefault(prop.Default); ok {
			s.Computed = false
			s.Default = boolDefault
		}
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

	if f := getKeyDiffSuppressFunc(key); f != nil {
		s.DiffSuppressFunc = f
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
		s, err := getSchema(key, prop)
		if err != nil {
			return nil, err
		}
		sMap[key] = s
	}
	return sMap, nil
}

func panicMessage(dbType upcloud.ManagedDatabaseServiceType, step string, err error) string {
	return fmt.Sprintf(`Could not generate %s properties %s. This is a bug in the provider. Please create an issue in https://github.com/UpCloudLtd/terraform-provider-upcloud/issues.

Error: %s`, dbType, step, err.Error())
}

func GetSchemaMap(dbType upcloud.ManagedDatabaseServiceType) map[string]*schema.Schema {
	p := GetProperties(dbType)

	s, err := getSchemaMap(p)
	if err != nil {
		panic(panicMessage(dbType, "schema", err))
	}
	return s
}
