package validator

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

var _ validator.String = isDomainNameValidator{}

// isDomainNameValidator validates that the value of an int64 attribute is divisible by a given divisor.
type isDomainNameValidator struct {
	divisor int64
}

// Description describes the validation.
func (v isDomainNameValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be divisible by %d", v.divisor)
}

// MarkdownDescription describes the validation in Markdown.
func (v isDomainNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v isDomainNameValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	err := ValidateDomainName(request.ConfigValue.ValueString())
	if err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			request.ConfigValue.ValueString(),
		))
	}
}

// IsDomainName returns an AttributeValidator to validate that the value is divisible by the given divisor.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func IsDomainName() validator.String {
	return isDomainNameValidator{}
}

func ValidateDomainName(name string) error {
	const (
		minLen      int = 1
		maxLen      int = 253
		labelMaxLen int = 63
	)
	l := len(name)

	if l > maxLen || l < minLen {
		return fmt.Errorf("%s length %d is not in the range %d - %d", name, l, minLen, maxLen)
	}

	if name[0] == '.' || name[0] == '-' {
		return fmt.Errorf("%s starts with dot or hyphen", name)
	}

	if name[l-1] == '.' || name[l-1] == '-' {
		return fmt.Errorf("%s ends with dot or hyphen", name)
	}

	last := byte('.')
	nonNumeric := false // true once we've seen a letter or hyphen (either one is required)
	labelLen := 0

	for i := 0; i < l; i++ {
		c := name[i]
		switch {
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			nonNumeric = true
			labelLen++
		case '0' <= c && c <= '9':
			labelLen++
		case c == '-':
			if last == '.' {
				return fmt.Errorf("'%s' character before hyphen cannot be dot", name[0:i+1])
			}
			labelLen++
			nonNumeric = true
		case c == '.':
			if last == '.' || last == '-' {
				return fmt.Errorf("'%s' character before dot cannot be dot or hyphen", name[0:i+1])
			}
			if labelLen > labelMaxLen || labelLen == 0 {
				return fmt.Errorf("'%s' label is not in the range %d - %d", name[0:i+1], minLen, labelMaxLen)
			}
			labelLen = 0
		default:
			return fmt.Errorf("%s contains illegal characters", name)
		}
		last = c
	}

	if labelLen > labelMaxLen {
		return fmt.Errorf("%s label is not in the range %d - %d", name, minLen, labelMaxLen)
	}

	if !nonNumeric {
		return fmt.Errorf("%s contains only numeric labels", name)
	}

	return nil
}

func ValidateDomainNameDiag(val interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	name, ok := val.(string)

	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Validation failed due to wrong type",
			Detail:        "expected value to be a string",
			AttributePath: path,
		})
		return diags
	}

	err := ValidateDomainName(name)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Validation failed",
			Detail:        err.Error(),
			AttributePath: path,
		})
	}

	return diags
}
