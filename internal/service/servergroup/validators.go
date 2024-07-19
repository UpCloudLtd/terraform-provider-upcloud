package servergroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Bool = trackMembersValidator{}

type trackMembersValidator struct{}

func (v trackMembersValidator) Description(_ context.Context) string {
	return "Validates that track_members is not set to false when members are set"
}

func (v trackMembersValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v trackMembersValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	var data serverGroupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !req.ConfigValue.IsNull() && !req.ConfigValue.ValueBool() && !data.Members.IsNull() {
		resp.Diagnostics.AddError("Invalid track_members value", "track_members can not be set to false when members set is not empty")
	}
}
