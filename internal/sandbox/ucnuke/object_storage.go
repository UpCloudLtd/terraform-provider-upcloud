package ucnuke

import (
	"context"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
)

func deleteObjectStorage(ctx context.Context, svc *service.Service, pk string) error {
	logf("deleting object storage id %s", pk)
	if _, ok := ctx.Deadline(); !ok {
		return errContextDeadlineNotSet
	}
	return errorIfResourceExists(svc.DeleteObjectStorage(ctx, &request.DeleteObjectStorageRequest{UUID: pk}))
}
