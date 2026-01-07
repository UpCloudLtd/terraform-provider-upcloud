package database

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database/properties"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWaitServiceNameToPropagate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	name := "upcloud.com"
	if err := waitServiceNameToPropagate(ctx, name); err != nil {
		t.Errorf("waitServiceNameToPropagate failed %+v", err)
	}
}

func TestWaitServiceNameToPropagateContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()
	name := "upcloud.com"
	if err := waitServiceNameToPropagate(ctx, name); err == nil {
		d, _ := ctx.Deadline()
		t.Errorf("waitServiceNameToPropagate failed didn't timeout before deadline %s", d.Format(time.RFC3339))
	}
}

func TestSetPropertyValue(t *testing.T) {
	ctx := context.Background()
	emptyListValue, _ := types.ListValueFrom(ctx, types.StringType, []types.String{})

	var emptyList any
	_ = json.Unmarshal([]byte("[]"), &emptyList)

	tests := []struct {
		name     string
		key      string
		value    any
		plan     attr.Value
		prop     upcloud.ManagedDatabaseServiceProperty
		expected any
	}{
		{
			name:  "empty ip_filter",
			key:   "ip_filter",
			value: emptyList,
			plan:  emptyListValue,
			prop: upcloud.ManagedDatabaseServiceProperty{
				Type: "array",
			},
			expected: emptyList,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, _ := test.plan.ToTerraformValue(ctx)
			processed, err := ignorePropChange(test.value, v, test.key, test.prop)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, processed)

			value, _ := properties.NativeToValue(ctx, processed, test.prop)
			assert.True(t, test.plan.Equal(value), "%s != %s", test.plan.String(), value.String())
		})
	}
}
