package server

import (
	"context"

	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud/service"
)

func addServerToGroup(ctx context.Context, service *service.Service, serverUUID, groupUUID string) error {
	if groupUUID == "" {
		return nil
	}

	return service.AddServerToServerGroup(ctx, &request.AddServerToServerGroupRequest{
		UUID:       groupUUID,
		ServerUUID: serverUUID,
	})
}

func removeServerFromGroup(ctx context.Context, service *service.Service, serverUUID, groupUUID string) error {
	if groupUUID == "" {
		return nil
	}

	return service.RemoveServerFromServerGroup(ctx, &request.RemoveServerFromServerGroupRequest{
		UUID:       groupUUID,
		ServerUUID: serverUUID,
	})
}
