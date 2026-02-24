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

	// ipList builds the []any slice that the API returns for an ip_filter value.
	ipList := func(ips ...string) any {
		list := make([]any, len(ips))
		for i, ip := range ips {
			list[i] = ip
		}
		return list
	}

	// planList builds the Terraform plan attr.Value for an ip_filter value.
	planList := func(ips ...string) attr.Value {
		strs := make([]types.String, len(ips))
		for i, ip := range ips {
			strs[i] = types.StringValue(ip)
		}
		v, _ := types.ListValueFrom(ctx, types.StringType, strs)
		return v
	}

	arrayProp := upcloud.ManagedDatabaseServiceProperty{Type: "array"}

	tests := []struct {
		name            string
		key             string
		value           any
		plan            attr.Value
		prop            upcloud.ManagedDatabaseServiceProperty
		expected        any
		wantMatchesPlan bool
	}{
		{
			name:            "empty ip_filter",
			key:             "ip_filter",
			value:           emptyList,
			plan:            emptyListValue,
			prop:            arrayProp,
			expected:        emptyList,
			wantMatchesPlan: true,
		},
		{
			name:            "ip_filter matches plan exactly",
			key:             "ip_filter",
			value:           ipList("1.2.3.4/32", "5.6.7.8/32"),
			plan:            planList("1.2.3.4/32", "5.6.7.8/32"),
			prop:            arrayProp,
			expected:        ipList("1.2.3.4/32", "5.6.7.8/32"),
			wantMatchesPlan: true,
		},
		{
			// The API adds a /32 suffix to bare IP addresses; the provider should
			// normalise this away so Terraform does not show a spurious diff.
			name:            "ip_filter api adds /32 suffix",
			key:             "ip_filter",
			value:           ipList("1.2.3.4/32"),
			plan:            planList("1.2.3.4"),
			prop:            arrayProp,
			expected:        ipList("1.2.3.4"),
			wantMatchesPlan: true,
		},
		{
			// An IP added via the web UI is present in the actual state but not in
			// the plan. The provider should surface the real state so that Terraform
			// can plan to remove the extra entry, rather than crashing.
			name:            "ip_filter actual has extra ip added outside terraform",
			key:             "ip_filter",
			value:           ipList("1.2.3.4/32", "5.6.7.8/32"),
			plan:            planList("1.2.3.4/32"),
			prop:            arrayProp,
			expected:        ipList("1.2.3.4/32", "5.6.7.8/32"),
			wantMatchesPlan: false,
		},
		{
			// An IP present in the plan was removed outside Terraform. The provider
			// should surface the real state so Terraform can plan to re-add it.
			name:            "ip_filter actual has fewer ips than plan",
			key:             "ip_filter",
			value:           ipList("1.2.3.4/32"),
			plan:            planList("1.2.3.4/32", "5.6.7.8/32"),
			prop:            arrayProp,
			expected:        ipList("1.2.3.4/32"),
			wantMatchesPlan: false,
		},
		{
			// An IP in the actual state differs from the planned IP. The provider
			// should surface the real state so Terraform can plan to correct it.
			name:            "ip_filter actual has different ip than plan",
			key:             "ip_filter",
			value:           ipList("9.9.9.9/32"),
			plan:            planList("1.2.3.4/32"),
			prop:            arrayProp,
			expected:        ipList("9.9.9.9/32"),
			wantMatchesPlan: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, _ := test.plan.ToTerraformValue(ctx)
			processed, err := ignorePropChange(test.value, v, test.key, test.prop)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, processed)

			value, _ := properties.NativeToValue(ctx, processed, test.prop)
			if test.wantMatchesPlan {
				assert.True(t, test.plan.Equal(value), "%s != %s", test.plan.String(), value.String())
			} else {
				assert.False(t, test.plan.Equal(value), "expected processed value to differ from plan, got: %s", value.String())
			}
		})
	}
}
