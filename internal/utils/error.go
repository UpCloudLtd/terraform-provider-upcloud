package utils

import (
	"errors"
	"net/http"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func HandleResourceError(resourceName string, d *schema.ResourceData, err error) diag.Diagnostics {
	var ucProb *upcloud.Problem
	if errors.As(err, &ucProb) && ucProb.Status == http.StatusNotFound {
		var diags diag.Diagnostics
		diags = append(diags, diagBindingRemovedWarningFromUpcloudProblem(ucProb, resourceName))
		d.SetId("")
		return diags
	}
	return diag.FromErr(err)
}

func ErrorDiagnosticDetail(err error) string {
	if err == nil {
		return ""
	}

	return "Error: " + err.Error()
}
