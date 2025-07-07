package properties

import (
	"context"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNativeToValue(t *testing.T) {
	tests := []struct {
		name     string
		native   any
		prop     upcloud.ManagedDatabaseServiceProperty
		expected attr.Value
	}{
		{
			name:   "string",
			native: "test-string",
			prop: upcloud.ManagedDatabaseServiceProperty{
				Type: "string",
			},
			expected: types.StringValue("test-string"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, diags := NativeToValue(context.Background(), test.native, test.prop)
			assert.False(t, diags.HasError())
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
		"prop_object": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"nested_string": types.StringType,
			},
		},
	}
	actual := PropsToAttributeTypes(props)
	assert.Equal(t, expected, actual)
}
