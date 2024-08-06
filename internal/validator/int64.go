package validator

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Int64 = divisibleByValidator{}

// divisibleByValidator validates that the value of an int64 attribute is divisible by a given divisor.
type divisibleByValidator struct {
	divisor int64
}

// Description describes the validation.
func (v divisibleByValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be divisible by %d", v.divisor)
}

// MarkdownDescription describes the validation in Markdown.
func (v divisibleByValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateInt64 validates.
func (v divisibleByValidator) ValidateInt64(ctx context.Context, request validator.Int64Request, response *validator.Int64Response) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if math.Mod(float64(request.ConfigValue.ValueInt64()), float64(v.divisor)) != 0 {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("%d", request.ConfigValue.ValueInt64()),
		))
	}
}

// DivisibleBy returns an AttributeValidator to validate that the value is divisible by the given divisor.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func DivisibleBy(divisor int64) validator.Int64 {
	if divisor == 0 {
		return nil
	}

	return divisibleByValidator{
		divisor: divisor,
	}
}
