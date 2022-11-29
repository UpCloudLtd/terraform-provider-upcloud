package ucnuke

import (
	"context"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
)

func deleteRouter(ctx context.Context, svc *service.Service, pk string) error {
	logf("deleting network router id %s", pk)
	if _, ok := ctx.Deadline(); !ok {
		return errContextDeadlineNotSet
	}
	for {
		r, err := svc.GetRouterDetails(ctx, &request.GetRouterDetailsRequest{UUID: pk})
		if err != nil {
			return errorIfResourceExists(err)
		}
		if len(r.AttachedNetworks) == 0 {
			break
		}
		logf("deleting network router id %s is waiting network(s): %s to deattach", pk, joinAttachedNetworksToIDs(r.AttachedNetworks, ", "))
		time.Sleep(time.Second * 5)
	}
	return errorIfResourceExists(svc.DeleteRouter(ctx, &request.DeleteRouterRequest{UUID: pk}))
}

func joinAttachedNetworksToIDs(nets upcloud.RouterNetworkSlice, sep string) string {
	s := make([]string, 0)
	for _, n := range nets {
		s = append(s, n.NetworkUUID)
	}
	return strings.Join(s, sep)
}
