package firewallruleset

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFirewallRulesetUpgradeStateV0DropsServerUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := firewallRulesetModelV0{
		ID:         types.StringValue("1200ecde-db95-4d1c-9133-6508f3232567"),
		Name:       types.StringValue("example"),
		Labels:     types.MapNull(types.StringType),
		Rules:      types.ListNull(ruleObjectType()),
		ServerUUID: types.StringValue("0077fa3d-32db-4b09-9f5f-30d9e9afb565"),
	}

	priorState := tfsdk.State{Schema: firewallRulesetSchemaV0()}
	diags := priorState.Set(ctx, prior)
	if diags.HasError() {
		t.Fatalf("set prior state: %v", diags)
	}

	req := resource.UpgradeStateRequest{State: &priorState}
	resp := resource.UpgradeStateResponse{
		State: tfsdk.State{Schema: firewallRulesetSchema()},
	}

	upgraders := (&firewallRulesetResource{}).UpgradeState(ctx)
	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatal("missing state upgrader for schema version 0")
	}
	upgrader.StateUpgrader(ctx, req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrade state: %v", resp.Diagnostics)
	}

	var upgraded firewallRulesetModel
	diags = resp.State.Get(ctx, &upgraded)
	if diags.HasError() {
		t.Fatalf("get upgraded state: %v", diags)
	}
	if upgraded.ID.ValueString() != prior.ID.ValueString() {
		t.Fatalf("id: got %q, want %q", upgraded.ID.ValueString(), prior.ID.ValueString())
	}
	if upgraded.Name.ValueString() != prior.Name.ValueString() {
		t.Fatalf("name: got %q, want %q", upgraded.Name.ValueString(), prior.Name.ValueString())
	}

	var leftover types.String
	diags = resp.State.GetAttribute(ctx, path.Root("server_uuid"), &leftover)
	if !diags.HasError() {
		t.Fatal("expected server_uuid to be absent from upgraded state")
	}
}
