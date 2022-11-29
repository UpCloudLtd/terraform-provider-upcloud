package ucnuke

import (
	"context"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
)

func deleteKubernetes(ctx context.Context, svc *service.Service, pk string) error {
	logf("deleting Kubernetes cluster id %s", pk)
	if _, ok := ctx.Deadline(); !ok {
		return errContextDeadlineNotSet
	}
	// TODO: test this more e.g. Do we need to wait other resource?
	return errorIfResourceExists(svc.DeleteKubernetesCluster(ctx, &request.DeleteKubernetesClusterRequest{UUID: pk}))
}
