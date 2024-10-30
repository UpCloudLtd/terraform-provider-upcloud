package managedobjectstorage

import (
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func unmarshalID(id string) (serviceUUID, domainName string, diags diag.Diagnostics) {
	err := utils.UnmarshalID(id, &serviceUUID, &domainName)
	if err != nil {
		diags.AddError(
			"Unable to unmarshal sub-resource ID",
			utils.ErrorDiagnosticDetail(err),
		)
	}
	return
}
