package main // import "github.com/UpCloudLtd/terraform-provider-upcloud"

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: upcloud.Provider,
	})
}
