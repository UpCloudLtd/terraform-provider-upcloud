package main // import "github.com/UpCloudLtd/terraform-provider-upcloud"

import (
	"flag"
	"log"

	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	factory, err := upcloud.NewProviderServerFactory()
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf5server.ServeOpt

	if debug {
		serveOpts = append(serveOpts, tf5server.WithManagedDebug())
	}

	err = tf5server.Serve(
		"registry.terraform.io/upcloudltd/upcloud",
		factory,
		serveOpts...,
	)
	if err != nil {
		log.Fatal(err)
	}
}
