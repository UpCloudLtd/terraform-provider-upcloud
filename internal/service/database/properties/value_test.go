package properties

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNativeToValue(t *testing.T) {
	objProp := upcloud.ManagedDatabaseServiceProperty{
		Type: "object",
		Properties: map[string]upcloud.ManagedDatabaseServiceProperty{
			"string": {
				Type: "string",
			},
			"integer": {
				Type: "integer",
			},
			"boolean": {
				Type: "boolean",
			},
		},
	}

	tests := []struct {
		name      string
		jsonInput []byte
		prop      upcloud.ManagedDatabaseServiceProperty
		expected  attr.Value
	}{
		{
			name:      "string",
			jsonInput: []byte(`"test-string"`),
			prop: upcloud.ManagedDatabaseServiceProperty{
				Type: "string",
			},
			expected: types.StringValue("test-string"),
		},
		{
			name:      "array",
			jsonInput: []byte(`["a","b","c"]`),
			prop: upcloud.ManagedDatabaseServiceProperty{
				Type: "array",
			},
			expected: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
				types.StringValue("c"),
			}),
		},
		{
			name:      "array (ip_filter to allow access from any IP)",
			jsonInput: []byte(`["0.0.0.0/0"]`),
			prop: upcloud.ManagedDatabaseServiceProperty{
				Type: "array",
			},
			expected: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("0.0.0.0/0"),
			}),
		},
		{
			name:      "object",
			jsonInput: []byte(`{"string": "abc", "integer": 123, "boolean": false}`),
			prop:      objProp,
			expected: types.ObjectValueMust(PropsToAttributeTypes(objProp.Properties), map[string]attr.Value{
				"string":  types.StringValue("abc"),
				"integer": types.Int64Value(123),
				"boolean": types.BoolValue(false),
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var input any
			err := json.Unmarshal(test.jsonInput, &input)
			assert.NoError(t, err)

			actual, diags := NativeToValue(context.Background(), input, test.prop)
			assert.False(t, diags.HasError())
			assert.Equal(t, test.expected, actual)
			assert.True(t, test.expected.Equal(actual))
		})
	}
}

func TestPropsToAttributeTypes(t *testing.T) {
	props := map[string]upcloud.ManagedDatabaseServiceProperty{
		"prop_string": {
			Type: "string",
		},
		"prop_integer": {
			Type: "integer",
		},
		"prop_boolean": {
			Type: "boolean",
		},
		"prop_number": {
			Type: "number",
		},
		"prop_array": {
			Type: "array",
		},
		"prop_object": {
			Type: "object",
			Properties: map[string]upcloud.ManagedDatabaseServiceProperty{
				"nested_string": {
					Type: "string",
				},
			},
		},
	}

	expected := map[string]attr.Type{
		"prop_string":  types.StringType,
		"prop_integer": types.Int64Type,
		"prop_boolean": types.BoolType,
		"prop_number":  types.Float64Type,
		"prop_array":   types.ListType{ElemType: types.StringType},
		"prop_object": types.ListType{ElemType: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"nested_string": types.StringType,
			},
		}},
	}
	actual := PropsToAttributeTypes(props)
	assert.Equal(t, expected, actual)
}
