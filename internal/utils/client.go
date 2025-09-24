package utils

import (
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func GetClientFromProviderData(providerData any) (client *service.Service, diags diag.Diagnostics) {
	if providerData == nil {
		return nil, diags
	}

	client, ok := providerData.(*service.Service)
	if !ok {
		diags.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected *service.Service, got: %T. Please report this issue to the provider developers.", providerData),
		)
	}

	return client, diags
}
