package utils

import (
	"context"
	"log"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
)

func VerifyServerStopped(ctx context.Context, stopRequest request.StopServerRequest, meta interface{}) error {
	if stopRequest.Timeout == 0 {
		stopRequest.Timeout = time.Minute * 2
	}
	if stopRequest.StopType == "" {
		stopRequest.StopType = upcloud.StopTypeSoft
	}

	client := meta.(*service.ServiceContext)
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
		log.Printf("[INFO] Stopping server (server UUID: %s)", stopRequest.UUID)
		_, err := client.StopServer(ctx, &stopRequest)
		if err != nil {
			return err
		}
		_, err = client.WaitForServerState(ctx, &request.WaitForServerStateRequest{
			UUID:         stopRequest.UUID,
			DesiredState: upcloud.ServerStateStopped,
			Timeout:      time.Minute * 5,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func VerifyServerStarted(ctx context.Context, startRequest request.StartServerRequest, meta interface{}) error {
	if startRequest.Timeout == 0 {
		startRequest.Timeout = time.Minute * 2
	}

	client := meta.(*service.ServiceContext)
	// Get current server state
	r := &request.GetServerDetailsRequest{
		UUID: startRequest.UUID,
	}
	server, err := client.GetServerDetails(ctx, r)
	if err != nil {
		return err
	}
	if server.State != upcloud.ServerStateStarted {
		log.Printf("[INFO] Starting server (server UUID: %s)", startRequest.UUID)
		_, err := client.StartServer(ctx, &startRequest)
		if err != nil {
			return err
		}
		_, err = client.WaitForServerState(ctx, &request.WaitForServerStateRequest{
			UUID:         startRequest.UUID,
			DesiredState: upcloud.ServerStateStarted,
			Timeout:      time.Minute * 5,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
