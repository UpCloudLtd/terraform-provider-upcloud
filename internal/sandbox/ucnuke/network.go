package ucnuke

import (
	"context"
	"net/http"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
)

func deleteNetwork(ctx context.Context, svc *service.Service, pk string) error {
	logf("deleting network id %s", pk)
	if _, ok := ctx.Deadline(); !ok {
		return errContextDeadlineNotSet
	}
	for {
		n, err := svc.GetNetworkDetails(ctx, &request.GetNetworkDetailsRequest{UUID: pk})
		if err != nil {
			if ucErr, ok := err.(*upcloud.Error); ok && ucErr.Status == http.StatusNotFound {
				return nil
			}
			return err
		}
		if len(n.Servers) == 0 {
			break
		}
		logf("deleting network id %s is waiting server(s): %s to deattach", pk, joinNetworkServerSliceToTitles(n.Servers, ", "))
		time.Sleep(time.Second * 5)
	}
	return errorIfResourceExists(svc.DeleteNetwork(ctx, &request.DeleteNetworkRequest{UUID: pk}))
}
