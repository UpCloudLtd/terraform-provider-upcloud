package properties

import (
	"context"
	"fmt"
	"math/big"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func listToValueMap(ctx context.Context, list types.List) (map[string]tftypes.Value, error) {
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

	propsMap, err := listToValueMap(ctx, list)
	if err != nil {
		return nil, err
	}

	for k, v := range propsMap {
		prop, ok := props[k]
		if !ok {
			continue
		}

		nativeValue, err := valueToNative(v, prop)
		if err != nil {
			return nil, err
		}

		res[upcloud.ManagedDatabasePropertyKey(k)] = nativeValue
	}

	return res, nil
}

func valueToNative(v tftypes.Value, prop upcloud.ManagedDatabaseServiceProperty) (any, error) {
	if v.IsNull() || !v.IsKnown() {
		return nil, nil
	}

	switch GetType(prop) {
	case "string":
		var s string
		err := v.As(&s)
		return s, err
	case "integer":
		var n big.Float
		err := v.As(&n)
		if err != nil {
			return nil, err
		}

		i, _ := n.Int64()
		return i, err
	case "number":
		var n big.Float
		err := v.As(&n)
		if err != nil {
			return nil, err
		}

		f, _ := n.Float64()
		return f, err
	case "boolean":
		var b bool
		err := v.As(&b)
		return b, err
	case "array":
		var l []tftypes.Value
		err := v.As(&l)
		if err != nil {
			return nil, err
		}

		res := make([]any, len(l))
		for i := range l {
			res[i], err = valueToNative(l[i], upcloud.ManagedDatabaseServiceProperty{Type: "string"})
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	case "object":
		m := map[string]tftypes.Value{}
		err := v.As(&m)
		if err != nil {
			return nil, err
		}

		res := make(map[string]any)
		for k := range m {
			res[k], err = valueToNative(m[k], prop.Properties[k])
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
	case "string":
		s, ok := v.(string)
		if !ok {
			return types.StringNull(), nil
		}
		return types.StringValue(s), nil
	case "integer":
		n, ok := v.(int64)
		if !ok {
			return types.Int64Null(), nil
		}
		return types.Int64Value(n), nil
	case "number":
		f, ok := v.(float64)
		if !ok {
			return types.Float64Null(), nil
		}
		return types.Float64Value(f), nil
	case "boolean":
		b, ok := v.(bool)
		if !ok {
			return types.BoolNull(), nil
		}
		return types.BoolValue(b), nil
	case "array":
		var l []types.String
		ss, ok := v.([]string)
		if !ok {
			return types.ListNull(types.StringType), nil
		}

		for _, s := range ss {
			l = append(l, types.StringValue(s))
		}
		return types.ListValueFrom(ctx, types.StringType, l)
	case "object":
		attrTypes := PropsToAttributeTypes(prop.Properties)

		m, ok := v.(map[string]any)
		if !ok {
			return types.ObjectNull(attrTypes), nil
		}

		o := make(map[string]attr.Value)
		for k, v := range m {
			o[k], d = NativeToValue(ctx, v, prop.Properties[k])
			diags.Append(d...)
		}

		return types.ObjectValue(attrTypes, o)
	default:
		diags.AddError("Unknown type", fmt.Sprintf(`unknown property value type "%s" for "%s"`, prop.Type, prop.Title))
		return nil, diags
	}
}

func PropsToAttributeTypes(props map[string]upcloud.ManagedDatabaseServiceProperty) map[string]attr.Type {
	attrTypes := make(map[string]attr.Type)
	for k, p := range props {
		attrTypes[k] = PropToAttributeType(p)
	}
	return attrTypes
}

func PropToAttributeType(prop upcloud.ManagedDatabaseServiceProperty) attr.Type {
	switch GetType(prop) {
	case "string":
		return types.StringType
	case "integer":
		return types.Int64Type
	case "number":
		return types.Float64Type
	case "boolean":
		return types.BoolType
	case "array":
		return types.ListType{ElemType: types.StringType}
	case "object":
		attrTypes := make(map[string]attr.Type)
		for k, p := range prop.Properties {
			attrTypes[k] = PropToAttributeType(p)
		}
		return types.ObjectType{AttrTypes: attrTypes}
	default:
		return nil
	}
}
