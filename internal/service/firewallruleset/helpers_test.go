package firewallruleset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerUUIDPlanDiagnostic(t *testing.T) {
	t.Parallel()

	serverA := types.StringValue("0077fa3d-32db-4b09-9f5f-30d9e9afb565")
	serverB := types.StringValue("190f56d8-4b3f-4a89-9e5f-320fbc3d17c8")

	tests := []struct {
		name     string
		creating bool
		state    types.String
		plan     types.String
		summary  string
	}{
		{
			name:     "create without server_uuid",
			creating: true,
			state:    types.StringNull(),
			plan:     types.StringNull(),
		},
		{
			name:     "create with unknown server_uuid",
			creating: true,
			state:    types.StringNull(),
			plan:     types.StringUnknown(),
		},
		{
			name:     "create with server_uuid",
			creating: true,
			state:    types.StringNull(),
			plan:     serverA,
			summary:  serverUUIDCreateSummary,
		},
		{
			name:     "update matching existing server_uuid",
			creating: false,
			state:    serverA,
			plan:     serverA,
		},
		{
			name:     "update omit server_uuid",
			creating: false,
			state:    serverA,
			plan:     types.StringNull(),
		},
		{
			name:     "update unknown server_uuid",
			creating: false,
			state:    serverA,
			plan:     types.StringUnknown(),
		},
		{
			name:     "update change server_uuid",
			creating: false,
			state:    serverA,
			plan:     serverB,
			summary:  serverUUIDChangeSummary,
		},
		{
			name:     "update add server_uuid",
			creating: false,
			state:    types.StringNull(),
			plan:     serverA,
			summary:  serverUUIDChangeSummary,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			summary, detail := serverUUIDPlanDiagnostic(tt.creating, tt.state, tt.plan)
			if summary != tt.summary {
				t.Fatalf("summary: got %q, want %q", summary, tt.summary)
			}
			if tt.summary == "" {
				if detail != "" {
					t.Fatalf("detail: got %q, want empty", detail)
				}
				return
			}
			if detail == "" {
				t.Fatal("expected non-empty detail")
			}
		})
	}
}
