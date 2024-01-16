package server

import (
	"context"

	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
)

func addServerToGroup(ctx context.Context, service *service.Service, serverUUID, groupUUID string) error {
	group, err := service.GetServerGroup(ctx, &request.GetServerGroupRequest{UUID: groupUUID})
	if err != nil {
		return err
	}

	members := group.Members
	members = append(members, serverUUID)

	_, err = service.ModifyServerGroup(ctx, &request.ModifyServerGroupRequest{UUID: groupUUID, Members: &members})
	return err
}

func removeServerFromGroup(ctx context.Context, service *service.Service, serverUUID, groupUUID string) error {
	group, err := service.GetServerGroup(ctx, &request.GetServerGroupRequest{UUID: groupUUID})
	if err != nil {
		return err
	}

	var members upcloud.ServerUUIDSlice
	for _, member := range group.Members {
		if member != serverUUID {
			members = append(members, member)
		}
	}

	_, err = service.ModifyServerGroup(ctx, &request.ModifyServerGroupRequest{UUID: groupUUID, Members: &members})
	return err
}
