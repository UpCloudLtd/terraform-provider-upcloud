package ucnuke

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
	"golang.org/x/sync/errgroup"
)

const (
	accountTimeout time.Duration = 30 * time.Minute

	PermissionTargetManagedKubernetes upcloud.PermissionTarget = "managed_kubernetes"
)

var errContextDeadlineNotSet = errors.New("context deadline not set")

type deleteFn map[upcloud.PermissionTarget]func(ctx context.Context, svc *service.Service, pk string) error

var targetHandlers = deleteFn{
	upcloud.PermissionTargetNetwork:             deleteNetwork,
	upcloud.PermissionTargetServer:              deleteServer,
	upcloud.PermissionTargetStorage:             deleteStorage,
	upcloud.PermissionTargetRouter:              deleteRouter,
	upcloud.PermissionTargetObjectStorage:       deleteObjectStorage,
	upcloud.PermissionTargetManagedDatabase:     deleteDatabase,
	upcloud.PermissionTargetManagedLoadbalancer: deleteLoadBalancer,
	PermissionTargetManagedKubernetes:           deleteKubernetes,
}

func Account(ctx context.Context, svc *service.Service) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), accountTimeout)
	defer cancel()
	permissions, err := svc.GetPermissions(ctxTimeout, &request.GetPermissionsRequest{})
	if err != nil {
		return err
	}
	if len(permissions) == 0 {
		return nil
	}
	g, ctx := errgroup.WithContext(ctxTimeout)
	for _, p := range permissions {
		if fn, ok := targetHandlers[p.TargetType]; ok {
			pk := p.TargetIdentifier
			g.Go(func() error {
				return fn(ctx, svc, pk)
			})
		}
		// TODO: handle tag_access and unknown/new resources
	}
	return g.Wait()
}

func logf(format string, v ...any) {
	// TODO: use env variable UCNUKE_LOG to control whether to print log entries.
	log.Printf(format, v...)
}

func errorIfResourceExists(err error) error {
	var serr *upcloud.Error
	if errors.As(err, &serr) {
		if serr.Status == http.StatusNotFound {
			return nil
		}
		return err
	}
	var perr *upcloud.Problem
	if errors.As(err, &perr) {
		if serr.Status == http.StatusNotFound {
			return nil
		}
		return err
	}
	return err
}
