package loadbalancer

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

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

func marshalID(components ...string) string {
	return strings.Join(components, "/")
}

func unmarshalID(id string, components ...*string) error {
	parts := strings.Split(id, "/")
	if len(parts) > len(components) {
		return fmt.Errorf("not enough components (%d) to unmarshal id '%s'", len(components), id)
	}
	for i, c := range parts {
		*components[i] = c
	}
	return nil
}

var validateNameDiagFunc = validation.ToDiagFunc(validation.StringMatch(
	regexp.MustCompile("^[a-zA-Z0-9_-]+$"),
	"should contain only alphanumeric characters, underscores and dashes",
))
