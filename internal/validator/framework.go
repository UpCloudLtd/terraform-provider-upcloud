package validator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type frameworkStringValidator struct {
	validateFunc schema.SchemaValidateFunc //nolint:staticcheck // Network validators use the deprecated schema.SchemaValidateFunc
}

func (v *frameworkStringValidator) Description(_ context.Context) string {
	return ""
}

func (v *frameworkStringValidator) MarkdownDescription(_ context.Context) string {
	return ""
}

func (v *frameworkStringValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	warnings, errors := v.validateFunc(req.ConfigValue.ValueString(), req.Path.String())

	for _, warning := range warnings {
		resp.Diagnostics = append(resp.Diagnostics, diag.NewWarningDiagnostic(warning, ""))
	}

	for _, err := range errors {
		resp.Diagnostics = append(resp.Diagnostics, diag.NewErrorDiagnostic(err.Error(), ""))
	}
}

var _ validator.String = &frameworkStringValidator{}

func NewFrameworkStringValidator(validate schema.SchemaValidateFunc) validator.String { //nolint:staticcheck // Network validators use the deprecated schema.SchemaValidateFunc
	return &frameworkStringValidator{
		validateFunc: validate,
	}
}
