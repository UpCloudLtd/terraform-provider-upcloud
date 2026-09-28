package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestServerDataSourceModelMatchesSchema(t *testing.T) {
	ctx := context.Background()
	var response datasource.SchemaResponse
	(&serverDataSource{}).Schema(ctx, datasource.SchemaRequest{}, &response)
	typ := response.Schema.Type().TerraformType(ctx)
	values := map[string]tftypes.Value{}
	for name, attrType := range typ.(tftypes.Object).AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	values["id"] = tftypes.NewValue(tftypes.String, "00000000-0000-0000-0000-000000000001")
	config := tfsdk.Config{Schema: response.Schema, Raw: tftypes.NewValue(typ, values)}
	var model serverDataSourceModel
	if diagnostics := config.Get(ctx, &model); diagnostics.HasError() {
		t.Fatalf("server data source cannot decode config: %v", diagnostics)
	}
}

func TestServerCreatePrivateFirewall(t *testing.T) {
	for _, tc := range []struct {
		name     string
		enabled  interface{}
		incoming interface{}
		outgoing interface{}
		fail     bool
	}{
		{"enabled", true, "drop", "accept", false},
		{"disabled", false, "accept", "drop", false},
		{"omitted", nil, nil, nil, false},
		{"unknown", tftypes.UnknownValue, tftypes.UnknownValue, tftypes.UnknownValue, false},
		{"actions_only", nil, "drop", "accept", false},
		{"modify_failure", true, "drop", "accept", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			const id = "00000000-0000-0000-0000-000000000001"
			details := map[string]interface{}{"uuid": id, "state": "started", "simple_backup": "no", "firewall": "on", "firewall_private": "off", "firewall_private_default_incoming_action": "accept", "firewall_private_default_outgoing_action": "accept"}
			var modifications []map[string]interface{}
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if req.Method == http.MethodPut {
					var body struct {
						Server map[string]interface{} `json:"server"`
					}
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Error(err)
						w.WriteHeader(400)
						return
					}
					modifications = append(modifications, body.Server)
					if tc.fail {
						w.WriteHeader(http.StatusBadRequest)
						_, _ = w.Write([]byte(`{"error":{"error_code":"INVALID_REQUEST","error_message":"injected modify failure"}}`))
						return
					}
					for k, v := range body.Server {
						details[k] = v
					}
				}
				if err := json.NewEncoder(w).Encode(map[string]interface{}{"server": details}); err != nil {
					t.Error(err)
				}
			}))
			defer api.Close()
			r := &serverResource{client: service.New(client.New("test", "test", client.WithBaseURL(api.URL)))}
			schema := r.getSchema(2)
			typ := schema.Type().TerraformType(ctx)
			values := map[string]tftypes.Value{}
			for name, attrType := range typ.(tftypes.Object).AttributeTypes {
				values[name] = tftypes.NewValue(attrType, nil)
			}
			values["firewall"] = tftypes.NewValue(tftypes.Bool, true)
			values["firewall_private"] = tftypes.NewValue(tftypes.Bool, tc.enabled)
			values["firewall_private_default_incoming_action"] = tftypes.NewValue(tftypes.String, tc.incoming)
			values["firewall_private_default_outgoing_action"] = tftypes.NewValue(tftypes.String, tc.outgoing)
			plan := tfsdk.Plan{Schema: schema, Raw: tftypes.NewValue(typ, values)}
			response := resource.CreateResponse{State: tfsdk.State{Schema: schema, Raw: tftypes.NewValue(typ, nil)}}
			r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
			require.Equal(t, tc.fail, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			var state serverModel
			require.False(t, response.State.Get(ctx, &state).HasError())
			require.Equal(t, id, state.ID.ValueString(), "created server must remain tracked even when modification fails")
			require.Equal(t, types.BoolValue(true), state.Firewall)
			if tc.name == "omitted" || tc.name == "unknown" {
				require.Empty(t, modifications)
				return
			}
			require.Len(t, modifications, 1)
			require.NotContains(t, modifications[0], "firewall", "private modification must not change the public firewall")
			if enabled, ok := tc.enabled.(bool); ok {
				want := "off"
				if enabled {
					want = "on"
				}
				require.Equal(t, want, modifications[0]["firewall_private"])
				if !tc.fail {
					require.Equal(t, types.BoolValue(enabled), state.FirewallPrivate)
				}
			} else {
				require.NotContains(t, modifications[0], "firewall_private")
			}
			require.Equal(t, tc.incoming, modifications[0]["firewall_private_default_incoming_action"])
			require.Equal(t, tc.outgoing, modifications[0]["firewall_private_default_outgoing_action"])
			if tc.fail {
				require.Equal(t, types.BoolValue(false), state.FirewallPrivate)
			} else {
				require.Equal(t, tc.incoming, state.FirewallPrivateDefaultIncomingAction.ValueString())
				require.Equal(t, tc.outgoing, state.FirewallPrivateDefaultOutgoingAction.ValueString())
			}
		})
	}
}
