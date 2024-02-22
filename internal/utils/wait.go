package utils

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func WaitForResourceToBeDeleted(ctx context.Context, svc *service.Service, getDetails func(context.Context, *service.Service, ...string) (map[string]interface{}, error), id ...string) error {
	const maxRetries int = 500

	for i := 0; i <= maxRetries; i++ {
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

	return fmt.Errorf("max retries (%d)reached while waiting for resource to be deleted", maxRetries)

}
