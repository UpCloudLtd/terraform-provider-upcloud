package ucnuke

import (
	"context"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
)

func deleteStorage(ctx context.Context, svc *service.Service, pk string) error {
	logf("deleting storage id %s", pk)
	if _, ok := ctx.Deadline(); !ok {
		return errContextDeadlineNotSet
	}
	for {
		s, err := svc.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{UUID: pk})
		if err != nil {
			return errorIfResourceExists(err)
		}
		if len(s.BackupUUIDs) > 0 {
			if err := deleteStorageBackups(ctx, svc, s.BackupUUIDs); err != nil {
				return err
			}
		}
		if len(s.ServerUUIDs) == 0 {
			break
		}
		logf("deleting storage id %s is waiting server(s): %s to deattach", pk, strings.Join(s.ServerUUIDs, ", "))
		time.Sleep(time.Second * 5)
	}
	return errorIfResourceExists(svc.DeleteStorage(ctx, &request.DeleteStorageRequest{UUID: pk}))
}

func deleteStorageBackups(ctx context.Context, svc *service.Service, pks upcloud.BackupUUIDSlice) error {
	for _, pk := range pks {
		if err := errorIfResourceExists(svc.DeleteStorage(ctx, &request.DeleteStorageRequest{UUID: pk})); err != nil {
			return err
		}
	}
	return nil
}
