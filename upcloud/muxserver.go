package upcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
)

func NewProviderServerFactory() (func() tfprotov5.ProviderServer, error) {
	providers := []func() tfprotov5.ProviderServer{
		providerserver.NewProtocol5(New()),
		Provider().GRPCProvider,
	}

	ctx := context.Background()
	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		return nil, err
	}

	return muxServer.ProviderServer, nil
}
