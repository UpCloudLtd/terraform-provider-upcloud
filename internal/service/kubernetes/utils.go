package kubernetes

import (
	"fmt"
	"net/http"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TODO: move these and duplicates under loadbalancer to utils package
func diagWarningFromUpcloudProblem(err *upcloud.Problem, details string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  err.Title,
		Detail:   details,
	}
}

func diagBindingRemovedWarningFromUpcloudProblem(err *upcloud.Problem, name string) diag.Diagnostic {
	return diagWarningFromUpcloudProblem(err,
		fmt.Sprintf("Binding to an existing remote object '%s' will be removed from the state. Next plan will include action to re-create the object if you choose to keep it in config.", name),
	)
}

func handleResourceError(resourceName string, d *schema.ResourceData, err error) diag.Diagnostics {
	if svcErr, ok := err.(*upcloud.Problem); ok && svcErr.Status == http.StatusNotFound {
		var diags diag.Diagnostics
		diags = append(diags, diagBindingRemovedWarningFromUpcloudProblem(svcErr, resourceName))
		d.SetId("")
		return diags
	}
	return diag.FromErr(err)
}
