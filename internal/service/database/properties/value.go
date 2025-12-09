package properties

import (
	"context"
	"fmt"
	"math/big"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const (
	propTypeString  string = "string"
	propTypeInteger string = "integer"
	propTypeNumber  string = "number"
	propTypeBoolean string = "boolean"
	propTypeArray   string = "array"
	propTypeObject  string = "object"
)

func ListToValueMap(ctx context.Context, list types.List) (map[string]tftypes.Value, error) {
	propsList := list.Elements()
	if len(propsList) == 0 {
		return nil, nil
	}
	propsValue, err := propsList[0].ToTerraformValue(ctx)
	if err != nil {
		return nil, err
	}

	props := make(map[string]tftypes.Value)
	err = propsValue.As(&props)
	if err != nil {
		return nil, err
	}
	return props, nil
}

func PlanToManagedDatabaseProperties(ctx context.Context, list types.List, props map[string]upcloud.ManagedDatabaseServiceProperty) (map[upcloud.ManagedDatabasePropertyKey]interface{}, error) {
	res := make(map[upcloud.ManagedDatabasePropertyKey]interface{})

	propsMap, err := ListToValueMap(ctx, list)
	if err != nil {
		return nil, err
	}

	for k, v := range propsMap {
		prop, ok := props[k]
		if !ok {
			continue
		}

		if v.IsNull() || !v.IsKnown() {
			continue
		}

		nativeValue, err := ValueToNative(v, prop)
		if err != nil {
			return nil, err
		}

		res[upcloud.ManagedDatabasePropertyKey(k)] = nativeValue
	}

	return res, nil
}

func ValueToNative(v tftypes.Value, prop upcloud.ManagedDatabaseServiceProperty) (any, error) {
	if v.IsNull() || !v.IsKnown() {
		return nil, nil
	}

	switch GetType(prop) {
	case propTypeString:
		var s string
		err := v.As(&s)
		return s, err
	case propTypeInteger:
		var n big.Float
		err := v.As(&n)
		if err != nil {
			return nil, err
		}

		i, _ := n.Int64()
		return i, err
	case propTypeNumber:
		var n big.Float
		err := v.As(&n)
		if err != nil {
			return nil, err
		}

		f, _ := n.Float64()
		return f, err
	case propTypeBoolean:
		var b bool
		err := v.As(&b)
		return b, err
	case propTypeArray:
		var l []tftypes.Value
		err := v.As(&l)
		if err != nil {
			return nil, err
		}

		res := make([]any, len(l))
		for i := range l {
			res[i], err = ValueToNative(l[i], upcloud.ManagedDatabaseServiceProperty{Type: "string"})
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	case propTypeObject:
		// We use nested lists for objects in schema, so convert to list first
		var l []tftypes.Value
		err := v.As(&l)
		if err != nil {
			return nil, err
		}

		if len(l) != 1 {
			return nil, nil
		}

		// After checking we have exactly one element, convert to map
		m := map[string]tftypes.Value{}
		err = l[0].As(&m)
		if err != nil {
			return nil, err
		}

		res := make(map[string]any)
		for k := range m {
			if m[k].IsNull() || !m[k].IsKnown() {
				continue
			}

			res[k], err = ValueToNative(m[k], prop.Properties[k])
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	default:
		return nil, fmt.Errorf(`unknown property value type "%s" for "%s"`, prop.Type, prop.Title)
	}
}

func NativeToValue(ctx context.Context, v any, prop upcloud.ManagedDatabaseServiceProperty) (attr.Value, diag.Diagnostics) {
	var d, diags diag.Diagnostics

	switch GetType(prop) {
	case propTypeString:
		s, ok := v.(string)
		if !ok {
			return types.StringNull(), nil
		}
		return types.StringValue(s), nil
	case propTypeInteger:
		// Numeric values in JSON are float64 by default
		f, ok := v.(float64)
		if !ok {
			return types.Int64Null(), nil
		}
		return types.Int64Value(int64(f)), nil
	case propTypeNumber:
		f, ok := v.(float64)
		if !ok {
			return types.Float64Null(), nil
		}
		return types.Float64Value(f), nil
	case propTypeBoolean:
		b, ok := v.(bool)
		if !ok {
			return types.BoolNull(), nil
		}
		return types.BoolValue(b), nil
	case propTypeArray:
		var l []types.String
		is, ok := v.([]any)
		if !ok {
			return types.ListNull(types.StringType), nil
		}

		for _, i := range is {
			s, ok := i.(string)
			if !ok {
				diags.AddError(
					"Failed to convert value to types.StringValue",
					fmt.Sprintf("Expected string, got %T", i),
				)
			} else {
				l = append(l, types.StringValue(s))
			}
		}

		val, d := types.ListValueFrom(ctx, types.StringType, l)
		diags.Append(d...)

		return val, diags
	case propTypeObject:
		attrTypes := PropsToAttributeTypes(prop.Properties)

		m, ok := v.(map[string]any)
		if !ok {
			return types.ObjectNull(attrTypes), nil
		}

		o := make(map[string]attr.Value)
		for k, v := range m {
			// Convert API keys to schema keys and recursively convert values
			o[SchemaKey(k)], d = NativeToValue(ctx, v, prop.Properties[k])
			diags.Append(d...)
		}

		// Add null values to omitted fields
		for k, v := range prop.Properties {
			schemaKey := SchemaKey(k)
			if _, ok := o[schemaKey]; !ok {
				o[schemaKey], d = NativeToValue(ctx, nil, v)
				diags.Append(d...)
			}
		}

		return types.ObjectValue(attrTypes, o)
	default:
		diags.AddError("Unknown type", fmt.Sprintf(`unknown property value type "%s" for "%s"`, prop.Type, prop.Title))
		return nil, diags
	}
}

func ValueToAttrValue(ctx context.Context, value tftypes.Value, prop upcloud.ManagedDatabaseServiceProperty) (attr.Value, diag.Diagnostics) {
	var d, diags diag.Diagnostics

	native, err := ValueToNative(value, prop)
	if err != nil {
		diags.AddError(
			"Failed to convert tftypes.Value to native Go value",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}

	attrValue, d := NativeToValue(ctx, native, prop)
	diags.Append(d...)

	return attrValue, diags
}

func ObjectValueAsList(v attr.Value, prop upcloud.ManagedDatabaseServiceProperty) (attr.Value, diag.Diagnostics) {
	// Pass through non object props
	if GetType(prop) != propTypeObject {
		return v, nil
	}

	// Use null list for null objects
	if v.IsNull() {
		return types.ListNull(PropToAttributeType(prop)), nil
	}

	return types.ListValue(PropToAttributeType(prop), []attr.Value{v})
}

func PropsToAttributeTypes(props map[string]upcloud.ManagedDatabaseServiceProperty) map[string]attr.Type {
	attrTypes := make(map[string]attr.Type)
	for k, p := range props {
		t := PropToAttributeType(p)
		// Wrap object types in a list because we use list-nested blocks for objects in schema.
		if GetType(p) == propTypeObject {
			t = types.ListType{ElemType: t}
		}
		attrTypes[k] = t
	}
	return attrTypes
}

func PropToAttributeType(prop upcloud.ManagedDatabaseServiceProperty) attr.Type {
	switch GetType(prop) {
	case propTypeString:
		return types.StringType
	case propTypeInteger:
		return types.Int64Type
	case propTypeNumber:
		return types.Float64Type
	case propTypeBoolean:
		return types.BoolType
	case propTypeArray:
		return types.ListType{ElemType: types.StringType}
	case propTypeObject:
		attrTypes := make(map[string]attr.Type)
		for k, p := range prop.Properties {
			attrTypes[k] = PropToAttributeType(p)
		}
		return types.ObjectType{AttrTypes: attrTypes}
	default:
		return nil
	}
}
