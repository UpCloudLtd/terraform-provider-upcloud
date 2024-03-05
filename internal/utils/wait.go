package utils

import (
	"context"
	"net/http"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func WaitForResourceToBeDeleted(ctx context.Context, svc *service.Service, getDetails func(context.Context, *service.Service, ...string) (map[string]interface{}, error), id ...string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			details, err := getDetails(ctx, svc, id...)
			if err != nil {
				if svcErr, ok := err.(*upcloud.Problem); ok && svcErr.Status == http.StatusNotFound {
					return nil
				}

				return err
			}

			tflog.Info(ctx, "waiting for resource to be deleted", details)
		}
		time.Sleep(5 * time.Second)
	}
}
