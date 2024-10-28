package upcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
)

func NewProviderServerFactory() (func() tfprotov6.ProviderServer, error) {
	ctx := context.Background()

	s, err := tf5to6server.UpgradeServer(ctx, Provider().GRPCProvider)
	if err != nil {
		return nil, err
	}

	providers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(New()),
		func() tfprotov6.ProviderServer {
			return s
		},
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		return nil, err
	}

	return muxServer.ProviderServer, nil
}
