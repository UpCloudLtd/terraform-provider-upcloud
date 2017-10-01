package main // import "github.com/vtorhonen/terraform-provider-upcloud"

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/vtorhonen/terraform-provider-upcloud/upcloud"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: upcloud.Provider})
}
