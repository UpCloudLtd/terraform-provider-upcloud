package utils

import (
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func DiagWarningFromUpcloudErr(err *upcloud.Error, details string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  err.ErrorMessage,
		Detail:   details,
	}
}

func DiagBindingRemovedWarningFromUpcloudErr(err *upcloud.Error, name string) diag.Diagnostic {
	return DiagWarningFromUpcloudErr(err,
		fmt.Sprintf("Binding to an existing remote object '%s' will be removed from the state. Next plan will include action to re-create the object if you choose to keep it in config.", name),
	)
}
