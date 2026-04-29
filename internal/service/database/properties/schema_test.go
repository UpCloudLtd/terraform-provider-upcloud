package properties

import (
	"context"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

var nonNullRaw = tftypes.NewValue(tftypes.String, "present")

func applyInt64PlanModifiers(t *testing.T, modifiers []planmodifier.Int64, req planmodifier.Int64Request) planmodifier.Int64Response {
	t.Helper()

	resp := planmodifier.Int64Response{PlanValue: req.PlanValue}
	for _, modifier := range modifiers {
		modifier.PlanModifyInt64(context.Background(), req, &resp)
		assert.False(t, resp.Diagnostics.HasError())
		req.PlanValue = resp.PlanValue
	}

	return resp
}

func applyBoolPlanModifiers(t *testing.T, modifiers []planmodifier.Bool, req planmodifier.BoolRequest) planmodifier.BoolResponse {
	t.Helper()

	resp := planmodifier.BoolResponse{PlanValue: req.PlanValue}
	for _, modifier := range modifiers {
		modifier.PlanModifyBool(context.Background(), req, &resp)
		assert.False(t, resp.Diagnostics.HasError())
		req.PlanValue = resp.PlanValue
	}

	return resp
}

func applyListPlanModifiers(t *testing.T, modifiers []planmodifier.List, req planmodifier.ListRequest) planmodifier.ListResponse {
	t.Helper()

	resp := planmodifier.ListResponse{PlanValue: req.PlanValue}
	for _, modifier := range modifiers {
		modifier.PlanModifyList(context.Background(), req, &resp)
		assert.False(t, resp.Diagnostics.HasError())
		req.PlanValue = resp.PlanValue
	}

	return resp
}

func TestCreateOnlyInt64PropertyPlanModifiers(t *testing.T) {
	s, err := getSchema("lower_case_table_names", upcloud.ManagedDatabaseServiceProperty{Type: "integer", CreateOnly: true})
	assert.NoError(t, err)

	attrSchema, ok := s.(schema.Int64Attribute)
	assert.True(t, ok)
	assert.Len(t, attrSchema.PlanModifiers, 2)

	t.Run("omitted config keeps state and does not require replace", func(t *testing.T) {
		resp := applyInt64PlanModifiers(t, attrSchema.PlanModifiers, planmodifier.Int64Request{
			ConfigValue: types.Int64Null(),
			PlanValue:   types.Int64Unknown(),
			StateValue:  types.Int64Value(0),
			Plan:        tfsdk.Plan{Raw: nonNullRaw},
			State:       tfsdk.State{Raw: nonNullRaw},
		})

		assert.False(t, resp.RequiresReplace)
		assert.Equal(t, types.Int64Value(0), resp.PlanValue)
	})

	t.Run("configured change requires replace", func(t *testing.T) {
		resp := applyInt64PlanModifiers(t, attrSchema.PlanModifiers, planmodifier.Int64Request{
			ConfigValue: types.Int64Value(1),
			PlanValue:   types.Int64Value(1),
			StateValue:  types.Int64Value(0),
			Plan:        tfsdk.Plan{Raw: nonNullRaw},
			State:       tfsdk.State{Raw: nonNullRaw},
		})

		assert.True(t, resp.RequiresReplace)
	})
}

func TestCreateOnlyBoolPropertyPlanModifiers(t *testing.T) {
	s, err := getSchema("feature_enabled", upcloud.ManagedDatabaseServiceProperty{Type: "boolean", CreateOnly: true})
	assert.NoError(t, err)

	attrSchema, ok := s.(schema.BoolAttribute)
	assert.True(t, ok)
	assert.Len(t, attrSchema.PlanModifiers, 2)

	resp := applyBoolPlanModifiers(t, attrSchema.PlanModifiers, planmodifier.BoolRequest{
		ConfigValue: types.BoolValue(false),
		PlanValue:   types.BoolValue(false),
		StateValue:  types.BoolValue(true),
		Plan:        tfsdk.Plan{Raw: nonNullRaw},
		State:       tfsdk.State{Raw: nonNullRaw},
	})

	assert.True(t, resp.RequiresReplace)
}

func TestCreateOnlyListPropertyPlanModifiers(t *testing.T) {
	s, err := getSchema("custom_repos", upcloud.ManagedDatabaseServiceProperty{Type: "array", CreateOnly: true})
	assert.NoError(t, err)

	attrSchema, ok := s.(schema.ListAttribute)
	assert.True(t, ok)
	assert.Len(t, attrSchema.PlanModifiers, 2)

	emptyList := types.ListValueMust(types.StringType, []attr.Value{})
	stateList := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("existing")})

	resp := applyListPlanModifiers(t, attrSchema.PlanModifiers, planmodifier.ListRequest{
		ConfigValue: emptyList,
		PlanValue:   emptyList,
		StateValue:  stateList,
		Plan:        tfsdk.Plan{Raw: nonNullRaw},
		State:       tfsdk.State{Raw: nonNullRaw},
	})

	assert.True(t, resp.RequiresReplace)
}
