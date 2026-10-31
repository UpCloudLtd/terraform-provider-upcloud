package database

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connectionTestSchema is a minimal stand-in for the managed database schema that
// contains only the attributes the connection details plan modifier inspects.
func connectionTestSchema(withValkeySSL bool) schema.Schema {
	nested := map[string]schema.Attribute{
		"ip_filter": schema.ListAttribute{ElementType: types.StringType, Optional: true},
	}
	if withValkeySSL {
		nested["valkey_ssl"] = schema.BoolAttribute{Optional: true, Computed: true}
	}

	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"powered":     schema.BoolAttribute{Optional: true, Computed: true},
			"service_uri": schema.StringAttribute{Computed: true},
		},
		Blocks: map[string]schema.Block{
			"properties": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{Attributes: nested},
			},
		},
	}
}

type connectionTestModel struct {
	Powered    types.Bool   `tfsdk:"powered"`
	ServiceURI types.String `tfsdk:"service_uri"`
	Properties types.List   `tfsdk:"properties"`
}

type connectionTestValues struct {
	powered   types.Bool
	ipFilter  []string
	valkeySSL attr.Value // nil leaves the key null; ignored when the schema has no valkey_ssl
	noProps   bool       // true omits the properties block entirely
}

func connectionTestRaw(t *testing.T, s schema.Schema, v connectionTestValues) tftypes.Value {
	t.Helper()

	objType, ok := s.Blocks["properties"].(schema.ListNestedBlock).NestedObject.Type().(types.ObjectType)
	require.True(t, ok)

	props := types.ListNull(objType)
	if !v.noProps {
		ips := make([]attr.Value, 0, len(v.ipFilter))
		for _, ip := range v.ipFilter {
			ips = append(ips, types.StringValue(ip))
		}
		vals := map[string]attr.Value{
			"ip_filter": types.ListValueMust(types.StringType, ips),
		}
		if _, ok := objType.AttrTypes["valkey_ssl"]; ok {
			if v.valkeySSL == nil {
				vals["valkey_ssl"] = types.BoolNull()
			} else {
				vals["valkey_ssl"] = v.valkeySSL
			}
		}
		props = types.ListValueMust(objType, []attr.Value{types.ObjectValueMust(objType.AttrTypes, vals)})
	}

	state := tfsdk.State{Schema: s}
	diags := state.Set(context.Background(), connectionTestModel{
		Powered:    v.powered,
		ServiceURI: types.StringValue("rediss://default:secret@host:11550"),
		Properties: props,
	})
	require.False(t, diags.HasError(), diags)

	return state.Raw
}

func TestConnectionDetailsUseStateForUnknown(t *testing.T) {
	t.Parallel()

	stateURI := types.StringValue("rediss://default:secret@host:11550")

	tests := []struct {
		name      string
		valkeySSL bool // whether the schema has a valkey_ssl property
		state     connectionTestValues
		plan      connectionTestValues
		wantKnown bool
	}{
		{
			name:      "ip_filter only change keeps state value",
			valkeySSL: true,
			state:     connectionTestValues{powered: types.BoolValue(true), ipFilter: []string{"10.0.0.1/32"}, valkeySSL: types.BoolValue(true)},
			plan:      connectionTestValues{powered: types.BoolValue(true), ipFilter: []string{"10.0.0.1/32", "10.0.0.2/32"}, valkeySSL: types.BoolValue(true)},
			wantKnown: true,
		},
		{
			name:      "schema without valkey_ssl keeps state value",
			valkeySSL: false,
			state:     connectionTestValues{powered: types.BoolValue(true), ipFilter: []string{"10.0.0.1/32"}},
			plan:      connectionTestValues{powered: types.BoolValue(true), ipFilter: []string{"10.0.0.1/32", "10.0.0.2/32"}},
			wantKnown: true,
		},
		{
			name:      "no properties block on either side keeps state value",
			valkeySSL: true,
			state:     connectionTestValues{powered: types.BoolValue(true), noProps: true},
			plan:      connectionTestValues{powered: types.BoolValue(true), noProps: true},
			wantKnown: true,
		},
		{
			name:      "properties block removed sends no change and keeps state value",
			valkeySSL: true,
			state:     connectionTestValues{powered: types.BoolValue(true), valkeySSL: types.BoolValue(true)},
			plan:      connectionTestValues{powered: types.BoolValue(true), noProps: true},
			wantKnown: true,
		},
		{
			name:      "valkey_ssl change leaves value unknown",
			valkeySSL: true,
			state:     connectionTestValues{powered: types.BoolValue(true), valkeySSL: types.BoolValue(false)},
			plan:      connectionTestValues{powered: types.BoolValue(true), valkeySSL: types.BoolValue(true)},
			wantKnown: false,
		},
		{
			name:      "unknown valkey_ssl leaves value unknown",
			valkeySSL: true,
			state:     connectionTestValues{powered: types.BoolValue(true), valkeySSL: types.BoolValue(true)},
			plan:      connectionTestValues{powered: types.BoolValue(true), valkeySSL: types.BoolUnknown()},
			wantKnown: false,
		},
		{
			name:      "properties block added with valkey_ssl leaves value unknown",
			valkeySSL: true,
			state:     connectionTestValues{powered: types.BoolValue(true), noProps: true},
			plan:      connectionTestValues{powered: types.BoolValue(true), valkeySSL: types.BoolValue(false)},
			wantKnown: false,
		},
		{
			name:      "powered change leaves value unknown",
			valkeySSL: true,
			state:     connectionTestValues{powered: types.BoolValue(true), valkeySSL: types.BoolValue(true)},
			plan:      connectionTestValues{powered: types.BoolValue(false), valkeySSL: types.BoolValue(true)},
			wantKnown: false,
		},
		{
			name:      "unknown powered leaves value unknown",
			valkeySSL: true,
			state:     connectionTestValues{powered: types.BoolValue(true), valkeySSL: types.BoolValue(true)},
			plan:      connectionTestValues{powered: types.BoolUnknown(), valkeySSL: types.BoolValue(true)},
			wantKnown: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := connectionTestSchema(tt.valkeySSL)
			req := planmodifier.StringRequest{
				Path:        path.Root("service_uri"),
				ConfigValue: types.StringNull(),
				PlanValue:   types.StringUnknown(),
				StateValue:  stateURI,
				State:       tfsdk.State{Schema: s, Raw: connectionTestRaw(t, s, tt.state)},
				Plan:        tfsdk.Plan{Schema: s, Raw: connectionTestRaw(t, s, tt.plan)},
			}
			resp := planmodifier.StringResponse{PlanValue: req.PlanValue}

			connectionDetailsUseStateForUnknown{}.PlanModifyString(context.Background(), req, &resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			if tt.wantKnown {
				assert.Equal(t, stateURI, resp.PlanValue)
			} else {
				assert.True(t, resp.PlanValue.IsUnknown(), "expected unknown, got %s", resp.PlanValue)
			}
		})
	}
}

func TestConnectionDetailsUseStateForUnknownOnCreate(t *testing.T) {
	t.Parallel()

	s := connectionTestSchema(true)
	req := planmodifier.StringRequest{
		Path:        path.Root("service_uri"),
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringUnknown(),
		StateValue:  types.StringNull(),
		State:       tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)},
		Plan:        tfsdk.Plan{Schema: s, Raw: connectionTestRaw(t, s, connectionTestValues{powered: types.BoolValue(true), valkeySSL: types.BoolValue(true)})},
	}
	resp := planmodifier.StringResponse{PlanValue: req.PlanValue}

	connectionDetailsUseStateForUnknown{}.PlanModifyString(context.Background(), req, &resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	assert.True(t, resp.PlanValue.IsUnknown())
}
