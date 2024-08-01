package utils

import (
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	sdkv2_diag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func diagWarningFromUpcloudProblem(err *upcloud.Problem, details string) sdkv2_diag.Diagnostic {
	return sdkv2_diag.Diagnostic{
		Severity: sdkv2_diag.Warning,
		Summary:  err.Title,
		Detail:   details,
	}
}

func diagBindingRemovedWarningFromUpcloudProblem(err *upcloud.Problem, name string) sdkv2_diag.Diagnostic {
	return diagWarningFromUpcloudProblem(err,
		fmt.Sprintf("Binding to an existing remote object '%s' will be removed from the state. Next plan will include action to re-create the object if you choose to keep it in config.", name),
	)
}

func AsSDKv2Diags(diags diag.Diagnostics) sdkv2_diag.Diagnostics {
	var sdkv2Diags sdkv2_diag.Diagnostics
	for _, d := range diags {
		var severity sdkv2_diag.Severity
		switch d.Severity() {
		case diag.SeverityWarning:
			severity = sdkv2_diag.Warning
		default:
			severity = sdkv2_diag.Error
		}

		sdkv2Diags = append(sdkv2Diags, sdkv2_diag.Diagnostic{
			Severity: severity,
			Summary:  d.Summary(),
			Detail:   d.Detail(),
		})
	}
	return sdkv2Diags
}
