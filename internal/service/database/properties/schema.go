package properties

import (
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
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func SchemaKey(key string) string {
	return strings.ReplaceAll(key, ".", "_")
}

func diffSuppressCreateOnlyProperty(_, _, _ string, d *sdkv2_schema.ResourceData) bool {
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

func getKeyDiffSuppressFunc(key string) sdkv2_schema.SchemaDiffSuppressFunc {
	switch key {
	case "ip_filter":
		return func(_, oldValue, newValue string, _ *sdkv2_schema.ResourceData) bool {
			return strings.TrimSuffix(oldValue, "/32") == strings.TrimSuffix(newValue, "/32")
		}
	default:
		return nil
	}
}

func isSensitive(key string) bool {
	return strings.Contains(key, "password")
}

func getPFSchema(key string, prop upcloud.ManagedDatabaseServiceProperty) (any, error) {
	switch GetType(prop) {
	case "string":
		s := schema.StringAttribute{
			Description:   getDescription(key, prop),
			Optional:      true,
			Computed:      true,
			Sensitive:     isSensitive(key),
			PlanModifiers: []planmodifier.String{},
			Validators:    []validator.String{},
		}

		if prop.CreateOnly {
			s.PlanModifiers = append(
				s.PlanModifiers,
				stringplanmodifier.RequiresReplace(),
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
	case "integer":
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
	case "number":
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
	case "boolean":
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
	case "array":
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
	case "object":
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
		return nil, fmt.Errorf(`unknown property value type "%s"`, prop.Type)
	}
}

func getSchema(key string, prop upcloud.ManagedDatabaseServiceProperty) (*sdkv2_schema.Schema, error) {
	s := sdkv2_schema.Schema{
		Description: getDescription(key, prop),
		Optional:    true,
		Computed:    true,
		Sensitive:   isSensitive(key),
	}

	if prop.CreateOnly {
		s.ForceNew = true
		s.DiffSuppressFunc = diffSuppressCreateOnlyProperty
	}

	validations := []sdkv2_schema.SchemaValidateFunc{} //nolint:staticcheck // Most of the validators still use the deprecated function signature.

	switch GetType(prop) {
	case "string":
		s.Type = sdkv2_schema.TypeString
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
		s.Type = sdkv2_schema.TypeInt

		if prop.Minimum != nil {
			validations = append(validations, validation.IntAtLeast(int(*prop.Minimum)))
		}
		if prop.Maximum != nil {
			if *prop.Maximum < float64(math.MaxInt) {
				validations = append(validations, validation.IntAtMost(int(*prop.Maximum)))
			}
		}
	case "number":
		s.Type = sdkv2_schema.TypeFloat

		if prop.Minimum != nil {
			validations = append(validations, validation.FloatAtLeast(*prop.Minimum))
		}
		if prop.Maximum != nil {
			validations = append(validations, validation.FloatAtMost(*prop.Maximum))
		}
	case "boolean":
		s.Type = sdkv2_schema.TypeBool

		if boolDefault, ok := booleanDefault(prop.Default); ok {
			s.Computed = false
			s.Default = boolDefault
		}
	case "array":
		s.Type = sdkv2_schema.TypeList
		s.Elem = &sdkv2_schema.Schema{
			Type: sdkv2_schema.TypeString,
		}
	case "object":
		nested, err := getSchemaMap(prop.Properties)
		if err != nil {
			return nil, err
		}

		s.Type = sdkv2_schema.TypeList
		s.MaxItems = 1
		s.Elem = &sdkv2_schema.Resource{Schema: nested}
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

func getNestedObject(props map[string]upcloud.ManagedDatabaseServiceProperty) (schema.NestedBlockObject, error) {
	o := schema.NestedBlockObject{}

	attributes := make(map[string]schema.Attribute)
	blocks := make(map[string]schema.Block)
	for key, prop := range props {
		s, err := getPFSchema(key, prop)
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

func getSchemaMap(props map[string]upcloud.ManagedDatabaseServiceProperty) (map[string]*sdkv2_schema.Schema, error) {
	sMap := make(map[string]*sdkv2_schema.Schema)
	for key, prop := range props {
		s, err := getSchema(key, prop)
		if err != nil {
			return nil, err
		}
		sMap[SchemaKey(key)] = s
	}
	return sMap, nil
}

func panicMessage(dbType upcloud.ManagedDatabaseServiceType, step string, err error) string {
	return fmt.Sprintf(`Could not generate %s properties %s. This is a bug in the provider. Please create an issue in https://github.com/UpCloudLtd/terraform-provider-upcloud/issues.

Error: %s`, dbType, step, err.Error())
}

func GetSchemaMap(dbType upcloud.ManagedDatabaseServiceType) map[string]*sdkv2_schema.Schema {
	p := GetProperties(dbType)

	s, err := getSchemaMap(p)
	if err != nil {
		panic(panicMessage(dbType, "schema", err))
	}
	return s
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
