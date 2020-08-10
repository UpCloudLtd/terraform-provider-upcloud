package main // import "github.com/UpCloudLtd/terraform-provider-upcloud"

import (
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return upcloud.Provider()
		},
	})
}
