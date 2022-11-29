package ucnuke

import (
	"context"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
)

func deleteLoadBalancer(ctx context.Context, svc *service.Service, pk string) error {
	logf("deleting load balancer id %s", pk)
	if _, ok := ctx.Deadline(); !ok {
		return errContextDeadlineNotSet
	}
	return errorIfResourceExists(svc.DeleteLoadBalancer(ctx, &request.DeleteLoadBalancerRequest{UUID: pk}))
}
