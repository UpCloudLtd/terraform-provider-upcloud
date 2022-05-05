package main // import "github.com/UpCloudLtd/terraform-provider-upcloud"

import (
	"flag"

	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	var debugMode bool
	var debugProviderAddr string

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.StringVar(&debugProviderAddr, "debug-provider-addr", "registry.terraform.io/upcloudltd/upcloud",
		"use same provider address as used in your configs")

	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return upcloud.Provider()
		},
		Debug: debugMode,
	}

	plugin.Serve(opts)
}
