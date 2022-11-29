package ucnuke

import (
	"context"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
)

func deleteDatabase(ctx context.Context, svc *service.Service, pk string) error {
	logf("deleting database id %s", pk)
	if _, ok := ctx.Deadline(); !ok {
		return errContextDeadlineNotSet
	}
	return errorIfResourceExists(svc.DeleteManagedDatabase(ctx, &request.DeleteManagedDatabaseRequest{UUID: pk}))
}
