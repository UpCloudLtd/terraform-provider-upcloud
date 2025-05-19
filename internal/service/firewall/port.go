package firewall

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = portValidator{}

// divisibleByValidator validates that the value of an int64 attribute is divisible by a given divisor.
type portValidator struct{}

// Description describes the validation.
func (v portValidator) Description(_ context.Context) string {
	return "value must be a valid port number"
}

// MarkdownDescription describes the validation in Markdown.
func (v portValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateInt64 validates.
func (v portValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	portStr := req.ConfigValue.ValueString()
	if portStr == "" {
		return
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		resp.Diagnostics.AddError(
			"Invalid port value",
			fmt.Sprintf("Expected a valid port number or an empty string, got: %s", portStr),
		)
		return
	}
}
