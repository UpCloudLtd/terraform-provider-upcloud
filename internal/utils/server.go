package utils

import (
	"context"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func VerifyServerStopped(ctx context.Context, stopRequest request.StopServerRequest, meta interface{}) error {
	if stopRequest.Timeout == 0 {
		stopRequest.Timeout = time.Minute * 2
	}
	if stopRequest.StopType == "" {
		stopRequest.StopType = upcloud.StopTypeSoft
	}

	client := meta.(*service.Service)
	// Get current server state
	r := &request.GetServerDetailsRequest{
		UUID: stopRequest.UUID,
	}
	server, err := client.GetServerDetails(ctx, r)
	if err != nil {
		return err
	}
	if server.State != upcloud.ServerStateStopped {
		// Soft stop with 2 minute timeout, after which hard stop occurs
		tflog.Info(ctx, "stopping server", map[string]interface{}{"uuid": stopRequest.UUID})
		_, err := client.StopServer(ctx, &stopRequest)
		if err != nil {
			return err
		}
		_, err = client.WaitForServerState(ctx, &request.WaitForServerStateRequest{
			UUID:         stopRequest.UUID,
			DesiredState: upcloud.ServerStateStopped,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func VerifyServerStarted(ctx context.Context, startRequest request.StartServerRequest, meta interface{}) error {
	client := meta.(*service.Service)
	// Get current server state
	r := &request.GetServerDetailsRequest{
		UUID: startRequest.UUID,
	}
	server, err := client.GetServerDetails(ctx, r)
	if err != nil {
		return err
	}
	if server.State != upcloud.ServerStateStarted {
		tflog.Info(ctx, "starting server", map[string]interface{}{"uuid": startRequest.UUID})
		_, err := client.StartServer(ctx, &startRequest)
		if err != nil {
			return err
		}
		_, err = client.WaitForServerState(ctx, &request.WaitForServerStateRequest{
			UUID:         startRequest.UUID,
			DesiredState: upcloud.ServerStateStarted,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
