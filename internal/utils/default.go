package utils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ defaults.Bool = StaticUnknown{}

// StaticUnknown sets an unknown default value. Terraform plugin framework should set null value in config to unknown value in plan, but if that does not happen, this can be used to force an unknown value to plan.
type StaticUnknown struct{}

func (d StaticUnknown) Description(_ context.Context) string {
	return "unknown default value to be used when Terraform does not set a null value to unknown during planning"
}

func (d StaticUnknown) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d StaticUnknown) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolUnknown()
}
