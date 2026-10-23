package server

import (
	"context"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
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

func TestBuildPrivateFirewallModifyRequest(t *testing.T) {
	const id = "00000000-0000-0000-0000-000000000001"
	tests := []struct {
		name       string
		enabled    types.Bool
		incoming   types.String
		outgoing   types.String
		want       request.ModifyServerRequest
		wantModify bool
	}{
		{
			name:    "enabled",
			enabled: types.BoolValue(true), incoming: types.StringValue("drop"), outgoing: types.StringValue("accept"),
			want:       request.ModifyServerRequest{UUID: id, FirewallPrivate: "on", FirewallPrivateDefaultIncomingAction: "drop", FirewallPrivateDefaultOutgoingAction: "accept"},
			wantModify: true,
		},
		{
			name:    "disabled",
			enabled: types.BoolValue(false), incoming: types.StringValue("accept"), outgoing: types.StringValue("drop"),
			want:       request.ModifyServerRequest{UUID: id, FirewallPrivate: "off", FirewallPrivateDefaultIncomingAction: "accept", FirewallPrivateDefaultOutgoingAction: "drop"},
			wantModify: true,
		},
		{
			name:    "omitted",
			enabled: types.BoolNull(), incoming: types.StringNull(), outgoing: types.StringNull(),
			want: request.ModifyServerRequest{UUID: id},
		},
		{
			name:    "unknown",
			enabled: types.BoolUnknown(), incoming: types.StringUnknown(), outgoing: types.StringUnknown(),
			want: request.ModifyServerRequest{UUID: id},
		},
		{
			name:    "actions only",
			enabled: types.BoolNull(), incoming: types.StringValue("drop"), outgoing: types.StringValue("accept"),
			want:       request.ModifyServerRequest{UUID: id, FirewallPrivateDefaultIncomingAction: "drop", FirewallPrivateDefaultOutgoingAction: "accept"},
			wantModify: true,
		},
		{
			name:    "incoming only with unknown enabled",
			enabled: types.BoolUnknown(), incoming: types.StringValue("drop"), outgoing: types.StringNull(),
			want:       request.ModifyServerRequest{UUID: id, FirewallPrivateDefaultIncomingAction: "drop"},
			wantModify: true,
		},
		{
			name:    "outgoing only with unknown incoming",
			enabled: types.BoolNull(), incoming: types.StringUnknown(), outgoing: types.StringValue("drop"),
			want:       request.ModifyServerRequest{UUID: id, FirewallPrivateDefaultOutgoingAction: "drop"},
			wantModify: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := serverModel{serverCommonModel: serverCommonModel{
				Firewall:                             types.BoolValue(true),
				FirewallPrivate:                      tt.enabled,
				FirewallPrivateDefaultIncomingAction: tt.incoming,
				FirewallPrivateDefaultOutgoingAction: tt.outgoing,
			}}
			got, modify := buildPrivateFirewallModifyRequest(id, data)
			require.Equal(t, tt.wantModify, modify)
			// Comparing the whole request also verifies that the public firewall is untouched.
			require.Equal(t, &tt.want, got)
		})
	}
}

func TestSetCommonValuesPrivateFirewall(t *testing.T) {
	tests := []struct {
		name     string
		enabled  string
		incoming string
		outgoing string
	}{
		{name: "enabled", enabled: "on", incoming: "drop", outgoing: "accept"},
		{name: "disabled", enabled: "off", incoming: "accept", outgoing: "drop"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data serverCommonModel
			var details upcloud.ServerDetails
			details.Firewall = "on"
			details.FirewallPrivate = tt.enabled
			details.FirewallPrivateDefaultIncomingAction = tt.incoming
			details.FirewallPrivateDefaultOutgoingAction = tt.outgoing
			diags := setCommonValues(context.Background(), &data, &details)
			require.False(t, diags.HasError(), "%v", diags)
			require.Equal(t, types.BoolValue(true), data.Firewall)
			require.Equal(t, types.BoolValue(tt.enabled == "on"), data.FirewallPrivate)
			require.Equal(t, types.StringValue(tt.incoming), data.FirewallPrivateDefaultIncomingAction)
			require.Equal(t, types.StringValue(tt.outgoing), data.FirewallPrivateDefaultOutgoingAction)
		})
	}
}
