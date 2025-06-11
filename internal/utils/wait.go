package utils

import (
	"context"
	"net/http"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ResourceDetails struct {
	ResourceType string
	Name         string
	State        string
	Running      bool
}

func (d *ResourceDetails) toMap() map[string]interface{} {
	return map[string]interface{}{
		"resource": d.ResourceType,
		"name":     d.Name,
		"state":    d.State,
		"running":  d.Running,
	}
}

func WaitForResourceToBeDeleted(ctx context.Context, svc *service.Service, getDetails func(context.Context, *service.Service, ...string) (*ResourceDetails, error), id ...string) error {
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

			tflog.Info(ctx, "waiting for resource to be deleted", details.toMap())
		}
		time.Sleep(5 * time.Second)
	}
}

func WaitForResourceToBeRunning(ctx context.Context, svc *service.Service, getDetails func(context.Context, *service.Service, ...string) (*ResourceDetails, error), id ...string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			details, err := getDetails(ctx, svc, id...)
			if err != nil {
				return err
			}

			if details.Running {
				return nil
			}

			tflog.Info(ctx, "waiting for resource to be running", details.toMap())
		}
		time.Sleep(5 * time.Second)
	}
}
