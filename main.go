package main // import "github.com/UpCloudLtd/terraform-provider-upcloud"

import (
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website
// IMPORTANT! the "go:generate" comments below are magic comments for "go generate" to function.

// Ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./upcloud/

// Run the docs generation tool
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return upcloud.Provider()
		},
	})
}
