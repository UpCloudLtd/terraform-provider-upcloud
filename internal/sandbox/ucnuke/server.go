package ucnuke

import (
	"context"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
)

func deleteServer(ctx context.Context, svc *service.Service, pk string) error {
	logf("deleting server id %s", pk)
	if _, ok := ctx.Deadline(); !ok {
		return errContextDeadlineNotSet
	}
	s, err := svc.GetServerDetails(ctx, &request.GetServerDetailsRequest{UUID: pk})
	if err != nil {
		return err
	}

	if s.State == upcloud.ServerStateMaintenance {
		// server is probably starting
		if _, err := svc.WaitForServerState(ctx, &request.WaitForServerStateRequest{
			UUID:         pk,
			DesiredState: upcloud.ServerStateStarted,
			Timeout:      time.Minute * 1,
		}); err != nil {
			return err
		}
	}
	if s.State != upcloud.ServerStateStopped {
		if _, err := svc.StopServer(ctx, &request.StopServerRequest{UUID: pk, StopType: "hard"}); err != nil {
			return errorIfResourceExists(err)
		}
		if _, err := svc.WaitForServerState(ctx, &request.WaitForServerStateRequest{
			UUID:         pk,
			DesiredState: upcloud.ServerStateStopped,
			Timeout:      time.Minute * 5,
		}); err != nil {
			return err
		}
	}
	return errorIfResourceExists(svc.DeleteServer(ctx, &request.DeleteServerRequest{UUID: pk}))
}

func joinNetworkServerSliceToTitles(servers upcloud.NetworkServerSlice, sep string) string {
	t := make([]string, 0)
	for _, s := range servers {
		t = append(t, s.ServerTitle)
	}
	return strings.Join(t, sep)
}
