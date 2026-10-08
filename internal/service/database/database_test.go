package database

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database/properties"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDatabaseTestClient(t *testing.T, handler http.HandlerFunc) *v9.ClientWithResponses {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := v9.New("ucat_testtoken", v9.WithBaseURL(server.URL))
	require.NoError(t, err)
	return client
}

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

func TestBuildManagedDatabaseRequestFromPlan(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		data          databaseCommonModel
		componentPlan *databasePlanModel
		want          map[string]any
		wantAbsent    []string
	}{
		"legacy plan": {
			data: databaseCommonModel{
				Name:                   types.StringValue("legacy-db"),
				Title:                  types.StringValue("Legacy database"),
				Type:                   types.StringValue(string(upcloud.ManagedDatabaseServiceTypeValkey)),
				Zone:                   types.StringValue("fi-hel1"),
				Plan:                   types.StringValue("1x1xCPU-2GB"),
				AdditionalDiskSpaceGiB: types.Int64Value(10),
				TerminationProtection:  types.BoolValue(true),
				Labels:                 types.MapNull(types.StringType),
				Network:                types.SetNull(types.ObjectType{}),
				Properties:             types.ListNull(types.ObjectType{}),
			},
			want: map[string]any{
				"hostname_prefix":           "legacy-db",
				"title":                     "Legacy database",
				"type":                      "valkey",
				"zone":                      "fi-hel1",
				"plan":                      "1x1xCPU-2GB",
				"additional_disk_space_gib": float64(10),
				"termination_protection":    true,
			},
			wantAbsent: []string{"plan_compute", "plan_node_count", "plan_storage_gib", "plan_backups"},
		},
		"componentized plan": {
			data: databaseCommonModel{
				Name:                   types.StringValue("component-db"),
				Title:                  types.StringValue("Component database"),
				Type:                   types.StringValue(string(upcloud.ManagedDatabaseServiceTypePostgreSQL)),
				Zone:                   types.StringValue("fi-hel1"),
				Plan:                   types.StringUnknown(),
				AdditionalDiskSpaceGiB: types.Int64Unknown(),
				TerminationProtection:  types.BoolValue(false),
				Labels:                 types.MapNull(types.StringType),
				Network:                types.SetNull(types.ObjectType{}),
				Properties:             types.ListNull(types.ObjectType{}),
			},
			componentPlan: &databasePlanModel{
				PlanCompute:    types.StringValue("rdb.standard.2CPU-8GB"),
				PlanNodeCount:  types.Int64Value(2),
				PlanStorageGiB: types.Int64Value(120),
				PlanBackups:    types.StringValue("regular"),
			},
			want: map[string]any{
				"hostname_prefix":        "component-db",
				"title":                  "Component database",
				"type":                   "pg",
				"zone":                   "fi-hel1",
				"termination_protection": false,
				"plan_compute":           "rdb.standard.2CPU-8GB",
				"plan_node_count":        float64(2),
				"plan_storage_gib":       float64(120),
				"plan_backups":           "regular",
			},
			wantAbsent: []string{"plan", "additional_disk_space_gib"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req, diags := buildManagedDatabaseRequestFromPlan(context.Background(), &test.data, test.componentPlan)
			assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			body, err := json.Marshal(req)
			assert.NoError(t, err)

			var actual map[string]any
			assert.NoError(t, json.Unmarshal(body, &actual))
			for key, value := range test.want {
				assert.Equal(t, value, actual[key], "unexpected %s", key)
			}
			for _, key := range test.wantAbsent {
				assert.NotContains(t, actual, key)
			}
		})
	}
}

func TestSetDatabaseValuesV9(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	componentsType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"component": types.StringType,
		"host":      types.StringType,
		"port":      types.Int64Type,
		"route":     types.StringType,
		"usage":     types.StringType,
	}}
	networkType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":   types.StringType,
		"type":   types.StringType,
		"family": types.StringType,
		"uuid":   types.StringType,
	}}
	nodeStatesType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":  types.StringType,
		"role":  types.StringType,
		"state": types.StringType,
	}}

	newModel := func() databaseCommonModel {
		return databaseCommonModel{
			Title:      types.StringValue("test database"),
			Labels:     types.MapNull(types.StringType),
			Components: types.ListNull(componentsType),
			Network:    types.SetNull(networkType),
			NodeStates: types.ListNull(nodeStatesType),
			Properties: types.ListNull(types.ObjectType{
				AttrTypes: properties.PropsToAttributeTypes(properties.GetProperties(upcloud.ManagedDatabaseServiceTypeValkey)),
			}),
		}
	}

	serviceUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	networkUUID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")

	strPtr := func(s string) *string { return &s }

	t.Run("full response", func(t *testing.T) {
		t.Parallel()

		var port int32 = 11599
		powered := true
		tp := false
		var additionalDisk int32 = 10
		db := &v9.DatabaseServiceInformationResponse{
			Uuid:                   &serviceUUID,
			Name:                   strPtr("test-db"),
			Title:                  strPtr("test database"),
			Type:                   strPtr("pg"),
			Zone:                   strPtr("fi-hel1"),
			Plan:                   strPtr("1x1xCPU-2GB-25GB"),
			State:                  strPtr("running"),
			Powered:                &powered,
			TerminationProtection:  &tp,
			AdditionalDiskSpaceGib: &additionalDisk,
			Maintenance: &v9.DatabaseMaintenanceWindowResponse{
				Dow:  strPtr("sunday"),
				Time: strPtr("05:00:00"),
			},
			ServiceUri: strPtr("postgresql://user:pass@host:11599/db"),
			ServiceUriParams: &map[string]interface{}{
				"dbname":   "defaultdb",
				"host":     "host.example.com",
				"password": "pass",
				"port":     float64(11599),
				"user":     "user",
				"ssl_mode": "require",
			},
			Labels: &[]v9.DatabaseLabelInformationResponse{
				{Key: strPtr("env"), Value: strPtr("test")},
			},
			Components: &[]v9.DatabaseServiceComponentResponse{
				{Component: strPtr("pg"), Host: strPtr("host.example.com"), Port: &port, Route: strPtr("dynamic"), Usage: strPtr("primary")},
			},
			Networks: &[]v9.DatabaseNetworkInformationDetailsResponse{
				{
					Family: (*v9.DatabaseNetworkInformationDetailsResponseFamily)(strPtr("IPv4")),
					Name:   strPtr("private-net"),
					Type:   (*v9.DatabaseNetworkInformationDetailsResponseType)(strPtr("private")),
					Uuid:   &networkUUID,
				},
			},
			NodeStates: &[]v9.DatabaseNodeStateResponse{
				{Name: strPtr("pg-1"), Role: strPtr("master"), State: strPtr("running")},
			},
		}

		data := newModel()
		diags := setDatabaseValues(ctx, &data, db, false)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

		assert.Equal(t, serviceUUID.String(), data.ID.ValueString())
		assert.Equal(t, "test-db", data.Name.ValueString())
		assert.Equal(t, "pg", data.Type.ValueString())
		assert.Equal(t, "running", data.State.ValueString())
		assert.True(t, data.Powered.ValueBool())
		assert.Equal(t, int64(10), data.AdditionalDiskSpaceGiB.ValueInt64())
		assert.Equal(t, "sunday", data.MaintenanceWindowDow.ValueString())
		assert.Equal(t, "host.example.com", data.ServiceHost.ValueString())
		assert.Equal(t, "11599", data.ServicePort.ValueString())
		assert.Equal(t, "defaultdb", data.PrimaryDatabase.ValueString())

		labels := map[string]string{}
		data.Labels.ElementsAs(ctx, &labels, false)
		assert.Equal(t, map[string]string{"env": "test"}, labels)

		var networks []databaseNetworkModel
		data.Network.ElementsAs(ctx, &networks, false)
		assert.Len(t, networks, 1)
		assert.Equal(t, "private", networks[0].Type.ValueString())
		assert.Equal(t, networkUUID.String(), networks[0].UUID.ValueString())

		var components []databaseComponentModel
		data.Components.ElementsAs(ctx, &components, false)
		assert.Len(t, components, 1)
		assert.Equal(t, int64(11599), components[0].Port.ValueInt64())

		var nodeStates []databaseNodeStateModel
		data.NodeStates.ElementsAs(ctx, &nodeStates, false)
		assert.Len(t, nodeStates, 1)
		assert.Equal(t, "master", nodeStates[0].Role.ValueString())
	})

	t.Run("normal sparse response preserves configured values", func(t *testing.T) {
		t.Parallel()

		data := newModel()
		data.ID = types.StringValue(serviceUUID.String())
		data.Plan = types.StringValue("1x1xCPU-2GB-25GB")
		data.AdditionalDiskSpaceGiB = types.Int64Value(10)
		data.MaintenanceWindowDow = types.StringValue("sunday")
		data.MaintenanceWindowTime = types.StringValue("05:00:00")
		data.TerminationProtection = types.BoolValue(true)
		data.Labels = types.MapValueMust(types.StringType, map[string]attr.Value{
			"env": types.StringValue("test"),
		})
		data.Network = types.SetValueMust(networkType, []attr.Value{
			types.ObjectValueMust(networkType.AttrTypes, map[string]attr.Value{
				"name":   types.StringValue("private-net"),
				"type":   types.StringValue("private"),
				"family": types.StringValue("IPv4"),
				"uuid":   types.StringValue(networkUUID.String()),
			}),
		})

		db := &v9.DatabaseServiceInformationResponse{
			Name:  strPtr("test-db"),
			Title: strPtr("test database"),
			Type:  strPtr("pg"),
			Zone:  strPtr("fi-hel1"),
			State: strPtr("running"),
		}

		diags := setDatabaseValues(ctx, &data, db, false)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

		assert.Equal(t, serviceUUID.String(), data.ID.ValueString())
		assert.Equal(t, "1x1xCPU-2GB-25GB", data.Plan.ValueString())
		assert.Equal(t, int64(10), data.AdditionalDiskSpaceGiB.ValueInt64())
		assert.Equal(t, "sunday", data.MaintenanceWindowDow.ValueString())
		assert.Equal(t, "05:00:00", data.MaintenanceWindowTime.ValueString())
		assert.True(t, data.TerminationProtection.ValueBool())
		assert.Equal(t, 1, len(data.Labels.Elements()))
		assert.Equal(t, 1, len(data.Network.Elements()))
	})

	t.Run("explicit empty collections clear state", func(t *testing.T) {
		t.Parallel()

		data := newModel()
		data.Labels = types.MapValueMust(types.StringType, map[string]attr.Value{
			"env": types.StringValue("test"),
		})
		data.Network = types.SetValueMust(networkType, []attr.Value{
			types.ObjectValueMust(networkType.AttrTypes, map[string]attr.Value{
				"name":   types.StringValue("private-net"),
				"type":   types.StringValue("private"),
				"family": types.StringValue("IPv4"),
				"uuid":   types.StringValue(networkUUID.String()),
			}),
		})
		db := &v9.DatabaseServiceInformationResponse{
			Uuid:     &serviceUUID,
			Name:     strPtr("test-db"),
			Title:    strPtr("test database"),
			Type:     strPtr("pg"),
			Zone:     strPtr("fi-hel1"),
			State:    strPtr("running"),
			Labels:   &[]v9.DatabaseLabelInformationResponse{},
			Networks: &[]v9.DatabaseNetworkInformationDetailsResponse{},
		}

		diags := setDatabaseValues(ctx, &data, db, false)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Empty(t, data.Labels.Elements())
		assert.Empty(t, data.Network.Elements())
	})

	t.Run("sparse response with nil fields", func(t *testing.T) {
		t.Parallel()

		db := &v9.DatabaseServiceInformationResponse{
			Uuid:  &serviceUUID,
			Name:  strPtr("test-db"),
			Title: strPtr("test database"),
			Type:  strPtr("valkey"),
			Zone:  strPtr("fi-hel1"),
			State: strPtr("stopped"),
		}

		data := newModel()
		diags := setDatabaseValues(ctx, &data, db, true)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

		assert.Equal(t, serviceUUID.String(), data.ID.ValueString())
		assert.Equal(t, "stopped", data.State.ValueString())
		assert.False(t, data.Powered.ValueBool())
		assert.True(t, data.Plan.IsNull())
		assert.True(t, data.ServiceURI.IsNull())
		assert.Equal(t, "", data.ServicePort.ValueString())
		assert.Equal(t, 0, len(data.Components.Elements()))
		assert.Equal(t, 0, len(data.Network.Elements()))
		assert.False(t, data.Properties.IsNull())
	})
}

func TestBuildManagedDatabaseModifyRequestFromPlan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baseModel := func() databaseCommonModel {
		return databaseCommonModel{
			ID:                     types.StringValue("550e8400-e29b-41d4-a716-446655440000"),
			Plan:                   types.StringValue("1x1xCPU-2GB-25GB"),
			Title:                  types.StringValue("test database"),
			Zone:                   types.StringValue("fi-hel1"),
			MaintenanceWindowDow:   types.StringValue("sunday"),
			MaintenanceWindowTime:  types.StringValue("05:00:00"),
			AdditionalDiskSpaceGiB: types.Int64Value(10),
			Labels: types.MapValueMust(types.StringType, map[string]attr.Value{
				"env": types.StringValue("test"),
			}),
			TerminationProtection: types.BoolValue(false),
			Properties:            types.ListNull(types.ObjectType{}),
			Network:               types.SetNull(types.ObjectType{}),
			Powered:               types.BoolValue(true),
		}
	}

	t.Run("component plan and title share one request", func(t *testing.T) {
		t.Parallel()

		state := baseModel()
		plan := baseModel()
		config := baseModel()
		plan.Title = types.StringValue("updated title")

		stateComponents := databasePlanModel{
			PlanCompute:    types.StringValue("rdb.standard.2CPU-8GB"),
			PlanNodeCount:  types.Int64Value(2),
			PlanStorageGiB: types.Int64Value(120),
			PlanBackups:    types.StringValue("regular"),
		}
		plannedComponents := stateComponents
		plannedComponents.PlanStorageGiB = types.Int64Value(140)
		configuredComponents := plannedComponents

		req, hasChanges, newVersion, diags := buildManagedDatabaseModifyRequestFromPlan(
			ctx,
			&state,
			&plan,
			&config,
			&stateComponents,
			&plannedComponents,
			&configuredComponents,
		)

		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.True(t, hasChanges)
		assert.Empty(t, newVersion)
		require.NotNil(t, req.Title)
		assert.Equal(t, "updated title", *req.Title)
		require.NotNil(t, req.PlanCompute)
		assert.Equal(t, "rdb.standard.2CPU-8GB", *req.PlanCompute)
		require.NotNil(t, req.PlanStorageGib)
		assert.Equal(t, 140, *req.PlanStorageGib)
	})

	t.Run("maintenance-only update is sent", func(t *testing.T) {
		t.Parallel()

		state := baseModel()
		plan := baseModel()
		config := baseModel()
		plan.MaintenanceWindowDow = types.StringValue("monday")
		plan.MaintenanceWindowTime = types.StringValue("06:00:00")

		req, hasChanges, _, diags := buildManagedDatabaseModifyRequestFromPlan(
			ctx,
			&state,
			&plan,
			&config,
			nil,
			nil,
			nil,
		)

		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.True(t, hasChanges)
		require.NotNil(t, req.Maintenance)
		assert.Equal(t, v9.DatabaseMaintenanceDow("monday"), req.Maintenance.Dow)
		assert.Equal(t, "06:00:00", req.Maintenance.Time)
	})

	t.Run("omitted maintenance is not sent after import", func(t *testing.T) {
		t.Parallel()

		state := baseModel()
		plan := baseModel()
		config := baseModel()
		plan.MaintenanceWindowDow = types.StringNull()
		plan.MaintenanceWindowTime = types.StringNull()
		config.MaintenanceWindowDow = types.StringNull()
		config.MaintenanceWindowTime = types.StringNull()

		req, hasChanges, _, diags := buildManagedDatabaseModifyRequestFromPlan(
			ctx,
			&state,
			&plan,
			&config,
			nil,
			nil,
			nil,
		)

		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.False(t, hasChanges)
		assert.Nil(t, req.Maintenance)
	})

	t.Run("unchanged plan does not send request", func(t *testing.T) {
		t.Parallel()

		state := baseModel()
		plan := baseModel()
		config := baseModel()

		_, hasChanges, newVersion, diags := buildManagedDatabaseModifyRequestFromPlan(
			ctx,
			&state,
			&plan,
			&config,
			nil,
			nil,
			nil,
		)

		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.False(t, hasChanges)
		assert.Empty(t, newVersion)
	})
}

func TestPollDatabase(t *testing.T) {
	t.Parallel()

	const databaseID = "550e8400-e29b-41d4-a716-446655440000"

	t.Run("pending then running", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32
		client := newDatabaseTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			state := "pending"
			if calls.Add(1) > 1 {
				state = "running"
			}
			_, _ = w.Write([]byte(`{"uuid":"` + databaseID + `","state":"` + state + `"}`))
		})

		db, diags := pollDatabaseWithInterval(
			context.Background(),
			client,
			databaseID,
			func(state string) bool { return state == databaseStateRunning },
			false,
			time.Millisecond,
		)

		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		require.NotNil(t, db)
		assert.Equal(t, databaseStateRunning, *db.State)
		assert.Equal(t, int32(2), calls.Load())
	})

	t.Run("terminal error includes state details", func(t *testing.T) {
		t.Parallel()

		client := newDatabaseTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"uuid":"` + databaseID + `","state":"error","state_error":{"setup":"provisioning failed"}}`))
		})

		_, diags := pollDatabaseWithInterval(
			context.Background(),
			client,
			databaseID,
			func(state string) bool { return state == databaseStateRunning },
			false,
			time.Millisecond,
		)

		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), "provisioning failed (setup)")
	})

	t.Run("not found is success only for deletion", func(t *testing.T) {
		t.Parallel()

		client := newDatabaseTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, `{"title":"not found"}`, http.StatusNotFound)
		})

		_, deleteDiags := pollDatabaseWithInterval(
			context.Background(),
			client,
			databaseID,
			func(string) bool { return false },
			true,
			time.Millisecond,
		)
		assert.False(t, deleteDiags.HasError(), "unexpected diagnostics: %v", deleteDiags)

		_, waitDiags := pollDatabaseWithInterval(
			context.Background(),
			client,
			databaseID,
			func(state string) bool { return state == databaseStateRunning },
			false,
			time.Millisecond,
		)
		require.True(t, waitDiags.HasError())
		assert.Contains(t, waitDiags[0].Detail(), "404")
	})

	t.Run("client error returns immediately", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32
		client := newDatabaseTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			http.Error(w, `{"title":"unauthorized"}`, http.StatusUnauthorized)
		})

		_, diags := pollDatabaseWithInterval(
			context.Background(),
			client,
			databaseID,
			func(state string) bool { return state == databaseStateRunning },
			false,
			time.Millisecond,
		)

		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), "401")
		assert.Equal(t, int32(1), calls.Load())
	})

	t.Run("transport error returns immediately", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		baseURL := server.URL
		server.Close()
		client, err := v9.New("ucat_testtoken", v9.WithBaseURL(baseURL))
		require.NoError(t, err)

		_, diags := pollDatabaseWithInterval(
			context.Background(),
			client,
			databaseID,
			func(state string) bool { return state == databaseStateRunning },
			false,
			time.Millisecond,
		)

		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Summary(), "Unable to get database details")
	})

	t.Run("transient server error recovers", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32
		client := newDatabaseTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if calls.Add(1) == 1 {
				http.Error(w, `{"title":"temporary"}`, http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(`{"uuid":"` + databaseID + `","state":"running"}`))
		})

		_, diags := pollDatabaseWithInterval(
			context.Background(),
			client,
			databaseID,
			func(state string) bool { return state == databaseStateRunning },
			false,
			time.Millisecond,
		)

		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Equal(t, int32(2), calls.Load())
	})

	t.Run("repeated server errors are bounded", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32
		client := newDatabaseTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			http.Error(w, `{"title":"temporary"}`, http.StatusInternalServerError)
		})

		_, diags := pollDatabaseWithInterval(
			context.Background(),
			client,
			databaseID,
			func(state string) bool { return state == databaseStateRunning },
			false,
			time.Millisecond,
		)

		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), "after 3 attempts")
		assert.Equal(t, int32(maxDatabasePollServerErrors), calls.Load())
	})

	t.Run("malformed successful response returns error", func(t *testing.T) {
		t.Parallel()

		client := newDatabaseTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"uuid":"` + databaseID + `"}`))
		})

		_, diags := pollDatabaseWithInterval(
			context.Background(),
			client,
			databaseID,
			func(state string) bool { return state == databaseStateRunning },
			false,
			time.Millisecond,
		)

		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), "without a state")
	})

	t.Run("context cancellation returns promptly", func(t *testing.T) {
		t.Parallel()

		client := newDatabaseTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"uuid":"` + databaseID + `","state":"pending"}`))
		})
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, diags := pollDatabaseWithInterval(
			ctx,
			client,
			databaseID,
			func(state string) bool { return state == databaseStateRunning },
			false,
			time.Millisecond,
		)

		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Summary(), "Context cancelled")
	})
}

func TestUpdateVersionAcceptsCreated(t *testing.T) {
	t.Parallel()

	const databaseID = "550e8400-e29b-41d4-a716-446655440000"
	var upgradeCalls atomic.Int32

	client := newDatabaseTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"uuid":"` + databaseID + `","state":"running"}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/tasks"):
			var body v9.CreateDatabaseTaskJSONRequestBody
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "upgrade_check", body.Operation)
			require.NotNil(t, body.UpgradeCheck)
			assert.Equal(t, "17", body.UpgradeCheck.TargetVersion)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"operation":"upgrade_check","success":true}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/upgrade"):
			upgradeCalls.Add(1)
			var body v9.UpgradeDatabaseJSONRequestBody
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "17", body.TargetVersion)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"uuid":"` + databaseID + `","state":"running"}`))
		default:
			http.NotFound(w, r)
		}
	})

	diags := updateVersion(context.Background(), databaseID, "17", true, client)
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Equal(t, int32(1), upgradeCalls.Load())
}

func TestWaitForDatabaseUpgradeCheck(t *testing.T) {
	t.Parallel()

	serviceUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	taskUUID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")

	t.Run("asynchronous check waits for success", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32
		client := newDatabaseTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/1.3/database/"+serviceUUID.String()+"/tasks/"+taskUUID.String(), r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			if calls.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"id":"` + taskUUID.String() + `","operation":"upgrade_check"}`))
				return
			}
			_, _ = w.Write([]byte(`{"id":"` + taskUUID.String() + `","operation":"upgrade_check","success":true}`))
		})
		task := &v9.DatabaseServiceTaskResponse{Id: &taskUUID}

		diags := waitForDatabaseUpgradeCheck(context.Background(), client, serviceUUID, task, time.Millisecond)

		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Equal(t, int32(2), calls.Load())
	})

	t.Run("failed check reports result", func(t *testing.T) {
		t.Parallel()

		success := false
		result := "extension is incompatible"
		task := &v9.DatabaseServiceTaskResponse{
			Success: &success,
			Result:  &result,
		}

		diags := waitForDatabaseUpgradeCheck(context.Background(), nil, serviceUUID, task, time.Millisecond)

		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), result)
	})
}

func TestUpdateDatabaseSkipsPowerChangeWhenWaitFails(t *testing.T) {
	t.Parallel()

	const databaseID = "550e8400-e29b-41d4-a716-446655440000"
	var patchCalls atomic.Int32

	client := newDatabaseTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"uuid":"` + databaseID + `","state":"error","state_error":{"power":"cannot change power state"}}`))
		case http.MethodPatch:
			patchCalls.Add(1)
			_, _ = w.Write([]byte(`{"uuid":"` + databaseID + `","state":"running"}`))
		default:
			http.NotFound(w, r)
		}
	})

	base := databaseCommonModel{
		ID:                     types.StringValue(databaseID),
		Plan:                   types.StringValue("1x1xCPU-2GB-25GB"),
		Title:                  types.StringValue("test database"),
		Zone:                   types.StringValue("fi-hel1"),
		MaintenanceWindowDow:   types.StringValue("sunday"),
		MaintenanceWindowTime:  types.StringValue("05:00:00"),
		AdditionalDiskSpaceGiB: types.Int64Value(10),
		Labels:                 types.MapNull(types.StringType),
		TerminationProtection:  types.BoolValue(false),
		Properties:             types.ListNull(types.ObjectType{}),
		Network:                types.SetNull(types.ObjectType{}),
		Powered:                types.BoolValue(true),
	}
	state := base
	plan := base
	config := base
	plan.Powered = types.BoolValue(false)

	_, _, diags := updateDatabase(
		context.Background(),
		&state,
		&plan,
		&config,
		nil,
		nil,
		nil,
		client,
	)

	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Detail(), "cannot change power state")
	assert.Equal(t, int32(0), patchCalls.Load())
}

func TestDeleteDatabaseAcceptsNoContentAndWaitsForNotFound(t *testing.T) {
	t.Parallel()

	const databaseID = "550e8400-e29b-41d4-a716-446655440000"
	var deleteCalls atomic.Int32
	var getCalls atomic.Int32

	client := newDatabaseTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			deleteCalls.Add(1)
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			getCalls.Add(1)
			http.Error(w, `{"title":"not found"}`, http.StatusNotFound)
		default:
			http.NotFound(w, r)
		}
	})

	diags := deleteDatabase(context.Background(), client, databaseID)
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Equal(t, int32(1), deleteCalls.Load())
	assert.Equal(t, int32(1), getCalls.Load())
}

func TestSetDatabasePlanComponents(t *testing.T) {
	t.Parallel()

	var components v9.DatabasePlanComponentsResponse
	err := json.Unmarshal([]byte(`{
		"compute": {"name": "rdb.standard.2CPU-8GB", "node_count": 2},
		"storage": {"total_gib": 120},
		"backups": {"name": "regular"}
	}`), &components)
	assert.NoError(t, err)

	data := databasePlanModel{
		PlanCompute:    types.StringNull(),
		PlanNodeCount:  types.Int64Null(),
		PlanStorageGiB: types.Int64Null(),
		PlanBackups:    types.StringNull(),
	}
	setDatabasePlanComponents(&data, &components)

	assert.Equal(t, types.StringValue("rdb.standard.2CPU-8GB"), data.PlanCompute)
	assert.Equal(t, types.Int64Value(2), data.PlanNodeCount)
	assert.Equal(t, types.Int64Value(120), data.PlanStorageGiB)
	assert.Equal(t, types.StringValue("regular"), data.PlanBackups)
}

func TestSetDatabasePlanComponentsPreservesMissingFields(t *testing.T) {
	t.Parallel()

	var components v9.DatabasePlanComponentsResponse
	require.NoError(t, json.Unmarshal([]byte(`{"storage":{"total_gib":140}}`), &components))
	data := databasePlanModel{
		PlanCompute:    types.StringValue("rdb.standard.2CPU-8GB"),
		PlanNodeCount:  types.Int64Value(2),
		PlanStorageGiB: types.Int64Value(120),
		PlanBackups:    types.StringValue("regular"),
	}

	setDatabasePlanComponents(&data, &components)

	assert.Equal(t, types.StringValue("rdb.standard.2CPU-8GB"), data.PlanCompute)
	assert.Equal(t, types.Int64Value(2), data.PlanNodeCount)
	assert.Equal(t, types.Int64Value(140), data.PlanStorageGiB)
	assert.Equal(t, types.StringValue("regular"), data.PlanBackups)
}

func TestResolveUnknownDatabasePlanComponents(t *testing.T) {
	t.Parallel()

	state := databasePlanModel{
		PlanCompute:    types.StringValue("1CPU-2GB"),
		PlanNodeCount:  types.Int64Value(1),
		PlanStorageGiB: types.Int64Value(25),
		PlanBackups:    types.StringNull(),
	}
	plan := databasePlanModel{
		PlanCompute:    types.StringUnknown(),
		PlanNodeCount:  types.Int64Unknown(),
		PlanStorageGiB: types.Int64Value(50),
		PlanBackups:    types.StringUnknown(),
	}

	resolveUnknownDatabasePlanComponents(&plan, &state)

	assert.Equal(t, state.PlanCompute, plan.PlanCompute)
	assert.Equal(t, state.PlanNodeCount, plan.PlanNodeCount)
	assert.Equal(t, types.Int64Value(50), plan.PlanStorageGiB)
	assert.Equal(t, types.StringNull(), plan.PlanBackups)

	var components v9.DatabasePlanComponentsResponse
	require.NoError(t, json.Unmarshal([]byte(`{"backups":{"name":"regular"}}`), &components))
	setDatabasePlanComponents(&plan, &components)
	assert.Equal(t, types.StringValue("regular"), plan.PlanBackups)
}

func TestDatabasePlanSchema(t *testing.T) {
	t.Parallel()

	var postgresResp resource.SchemaResponse
	(&postgresResource{}).Schema(context.Background(), resource.SchemaRequest{}, &postgresResp)
	assert.False(t, postgresResp.Diagnostics.HasError())

	plan := postgresResp.Schema.Attributes["plan"].(schema.StringAttribute)
	assert.True(t, plan.Optional)
	assert.True(t, plan.Computed)
	assert.NotEmpty(t, plan.DeprecationMessage)
	for _, name := range []string{"plan_compute", "plan_node_count", "plan_storage_gib", "plan_backups"} {
		assert.Contains(t, postgresResp.Schema.Attributes, name)
	}

	var valkeyResp resource.SchemaResponse
	(&valkeyResource{}).Schema(context.Background(), resource.SchemaRequest{}, &valkeyResp)
	assert.False(t, valkeyResp.Diagnostics.HasError())

	plan = valkeyResp.Schema.Attributes["plan"].(schema.StringAttribute)
	assert.True(t, plan.Required)
	assert.Empty(t, plan.DeprecationMessage)
	for _, name := range []string{"plan_compute", "plan_node_count", "plan_storage_gib", "plan_backups"} {
		assert.NotContains(t, valkeyResp.Schema.Attributes, name)
	}
}

func TestDatabaseComponentPlanChanged(t *testing.T) {
	t.Parallel()

	state := databasePlanModel{
		PlanCompute:    types.StringValue("rdb.standard.2CPU-8GB"),
		PlanNodeCount:  types.Int64Value(2),
		PlanStorageGiB: types.Int64Value(120),
		PlanBackups:    types.StringValue("regular"),
	}
	plan := state
	assert.False(t, databaseComponentPlanChanged(&state, &plan))

	plan.PlanStorageGiB = types.Int64Value(140)
	assert.True(t, databaseComponentPlanChanged(&state, &plan))
}
