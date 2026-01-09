package properties

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// GetType parses property type from the raw value, which might be either a string or a list of strings. If the type is a list of strings, first non-null value is returned.
func GetType(prop upcloud.ManagedDatabaseServiceProperty) string {
	raw := prop.Type
	if types, ok := raw.([]interface{}); ok {
		for _, t := range types {
			if t.(string) != "null" {
				return t.(string)
			}
		}
	}
	if t, ok := raw.(string); ok {
		return t
	}
	return ""
}

// GetKey converts key used in schema to key used in the API and properties info. E.g., "pressure_enabled" -> "pressure.enabled"
func GetKey(props map[string]upcloud.ManagedDatabaseServiceProperty, field string) string {
	if _, ok := props[field]; ok {
		return field
	}

	withDots := strings.ReplaceAll(field, "_", ".")
	if _, ok := props[withDots]; ok {
		return withDots
	}

	return ""
}

// SchemaKey converts key used in API to format supported in TF state. E.g., "pressure.enabled" -> "pressure_enabled"
func SchemaKey(key string) string {
	return strings.ReplaceAll(key, ".", "_")
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

func isSensitive(key string) bool {
	return strings.Contains(key, "password")
}

func getSchema(key string, prop upcloud.ManagedDatabaseServiceProperty) (any, error) {
	switch GetType(prop) {
	case propTypeString:
		s := schema.StringAttribute{
			Description:   getDescription(key, prop),
			Optional:      true,
			Computed:      true,
			Sensitive:     isSensitive(key),
			PlanModifiers: []planmodifier.String{},
			Validators:    []validator.String{},
		}

		if prop.CreateOnly {
			replaceIfDescription := "Do not require replace on import."
			s.PlanModifiers = append(
				s.PlanModifiers,
				stringplanmodifier.RequiresReplaceIf(
					func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
						if req.ConfigValue.IsNull() {
							resp.RequiresReplace = false
							return
						}
						resp.RequiresReplace = true
					},
					replaceIfDescription,
					replaceIfDescription,
				),
				stringplanmodifier.UseStateForUnknown(),
			)
		}

		if prop.MaxLength != 0 {
			s.Validators = append(s.Validators, stringvalidator.LengthBetween(prop.MinLength, prop.MaxLength))
		}
		if enum, hasEnum := stringSlice(prop.Enum); hasEnum {
			s.Validators = append(s.Validators, stringvalidator.OneOf(enum...))
		}
		if prop.Pattern != "" {
			if re, err := regexp.Compile(prop.Pattern); err == nil {
				s.Validators = append(s.Validators, stringvalidator.RegexMatches(re, fmt.Sprintf(`Must match "%s" pattern.`, prop.Pattern)))
			}
		}

		return s, nil
	case propTypeInteger:
		s := schema.Int64Attribute{
			Description:   getDescription(key, prop),
			Optional:      true,
			Computed:      true,
			Sensitive:     isSensitive(key),
			PlanModifiers: []planmodifier.Int64{},
			Validators:    []validator.Int64{},
		}

		if prop.CreateOnly {
			s.PlanModifiers = append(
				s.PlanModifiers,
				int64planmodifier.RequiresReplace(),
				int64planmodifier.UseStateForUnknown(),
			)
		}

		if prop.Minimum != nil {
			s.Validators = append(s.Validators, int64validator.AtLeast(int64(*prop.Minimum)))
		}
		if prop.Maximum != nil {
			if *prop.Maximum < float64(math.MaxInt) {
				s.Validators = append(s.Validators, int64validator.AtMost(int64(*prop.Maximum)))
			}
		}

		return s, nil
	case propTypeNumber:
		s := schema.Float64Attribute{
			Description:   getDescription(key, prop),
			Optional:      true,
			Computed:      true,
			Sensitive:     isSensitive(key),
			PlanModifiers: []planmodifier.Float64{},
			Validators:    []validator.Float64{},
		}

		if prop.CreateOnly {
			s.PlanModifiers = append(
				s.PlanModifiers,
				float64planmodifier.RequiresReplace(),
				float64planmodifier.UseStateForUnknown(),
			)
		}

		if prop.Minimum != nil {
			s.Validators = append(s.Validators, float64validator.AtLeast(*prop.Minimum))
		}
		if prop.Maximum != nil {
			s.Validators = append(s.Validators, float64validator.AtMost(*prop.Maximum))
		}

		return s, nil
	case propTypeBoolean:
		s := schema.BoolAttribute{
			Description:   getDescription(key, prop),
			Optional:      true,
			Computed:      true,
			Sensitive:     isSensitive(key),
			PlanModifiers: []planmodifier.Bool{},
		}

		if prop.CreateOnly {
			s.PlanModifiers = append(
				s.PlanModifiers,
				boolplanmodifier.RequiresReplace(),
				boolplanmodifier.UseStateForUnknown(),
			)
		}

		if boolDefault, ok := booleanDefault(prop.Default); ok {
			s.Default = booldefault.StaticBool(boolDefault)
		}

		return s, nil
	case propTypeArray:
		s := schema.ListAttribute{
			Description:   getDescription(key, prop),
			ElementType:   types.StringType,
			Optional:      true,
			Computed:      true,
			Sensitive:     isSensitive(key),
			PlanModifiers: []planmodifier.List{},
		}

		if prop.CreateOnly {
			s.PlanModifiers = append(
				s.PlanModifiers,
				listplanmodifier.RequiresReplace(),
				listplanmodifier.UseStateForUnknown(),
			)
		}

		return s, nil
	case propTypeObject:
		nested, err := getNestedObject(prop.Properties)
		if err != nil {
			return nil, err
		}

		s := schema.ListNestedBlock{
			Description:   getDescription(key, prop),
			NestedObject:  nested,
			PlanModifiers: []planmodifier.List{},
			Validators: []validator.List{
				listvalidator.SizeAtMost(1),
			},
		}

		if prop.CreateOnly {
			s.PlanModifiers = append(
				s.PlanModifiers,
				listplanmodifier.RequiresReplace(),
				listplanmodifier.UseStateForUnknown(),
			)
		}

		return s, nil
	default:
		return nil, fmt.Errorf(`unknown property value in %#v for key "%s"`, prop, key)
	}
}

func getNestedObject(props map[string]upcloud.ManagedDatabaseServiceProperty) (schema.NestedBlockObject, error) {
	o := schema.NestedBlockObject{}

	attributes := make(map[string]schema.Attribute)
	blocks := make(map[string]schema.Block)
	for key, prop := range props {
		s, err := getSchema(key, prop)
		if err != nil {
			return o, err
		}

		if block, ok := s.(schema.Block); ok {
			blocks[SchemaKey(key)] = block
		}
		if attribute, ok := s.(schema.Attribute); ok {
			attributes[SchemaKey(key)] = attribute
		}
	}

	o.Attributes = attributes
	o.Blocks = blocks
	return o, nil
}

func panicMessage(dbType upcloud.ManagedDatabaseServiceType, step string, err error) string {
	return fmt.Sprintf(`Could not generate %s properties %s. This is a bug in the provider. Please create an issue in https://github.com/UpCloudLtd/terraform-provider-upcloud/issues.

Error: %s`, dbType, step, err.Error())
}

func GetBlock(dbType upcloud.ManagedDatabaseServiceType) schema.Block {
	props := GetProperties(dbType)

	nested, err := getNestedObject(props)
	if err != nil {
		panic(panicMessage(dbType, "block", err))
	}
	block := schema.ListNestedBlock{
		MarkdownDescription: "Database engine properties.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: nested,
	}
	return block
}
