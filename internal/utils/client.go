package utils

import (
	"fmt"

	v9 "github.com/UpCloudLtd/upcloud-go-api-generated/pkg/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type ServiceWithV9Client struct {
	Service  *service.Service
	V9Client *v9.ClientWithResponses
}

func getServiceWithV9ClientFromProviderData(providerData any) (*ServiceWithV9Client, diag.Diagnostics) {
	var diags diag.Diagnostics

	if providerData == nil {
		return nil, diags
	}

	withV9, ok := providerData.(ServiceWithV9Client)
	if !ok {
		diags.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected ServiceWithV9Client, got: %T. Please report this issue to the provider developers.", providerData),
		)
	}

	return &withV9, diags
}

func GetClientFromProviderData(providerData any) (client *service.Service, diags diag.Diagnostics) {
	withV9, diags := getServiceWithV9ClientFromProviderData(providerData)
	if diags.HasError() || withV9 == nil {
		return nil, diags
	}

	return withV9.Service, diags
}

func GetV9ClientFromProviderData(providerData any) (*v9.ClientWithResponses, diag.Diagnostics) {
	withV9, diags := getServiceWithV9ClientFromProviderData(providerData)
	if diags.HasError() || withV9 == nil {
		return nil, diags
	}

	return withV9.V9Client, diags
}
