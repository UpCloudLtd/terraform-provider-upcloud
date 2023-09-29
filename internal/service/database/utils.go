package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUpCloudManagedDatabaseWaitState(
	ctx context.Context,
	id string,
	m interface{},
	timeout time.Duration,
	targetStates ...upcloud.ManagedDatabaseState,
) (*upcloud.ManagedDatabase, error) {
	client := m.(*service.Service)
	refresher := func() (result interface{}, state string, err error) {
		resp, err := client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: id})
		if err != nil {
			return nil, "", err
		}
		return resp, string(resp.State), nil
	}
	res, state, err := refresher()
	if err != nil {
		return nil, err
	}
	if len(targetStates) == 0 {
		return res.(*upcloud.ManagedDatabase), nil
	}
	for _, targetState := range targetStates {
		if upcloud.ManagedDatabaseState(state) == targetState {
			return res.(*upcloud.ManagedDatabase), nil
		}
	}
	states := make([]string, 0, len(targetStates))
	for _, targetState := range targetStates {
		states = append(states, string(targetState))
	}
	waiter := retry.StateChangeConf{
		Delay:      1 * time.Second,
		Refresh:    refresher,
		Target:     states,
		Timeout:    timeout,
		MinTimeout: 2 * time.Second,
	}
	res, err = waiter.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}
	return res.(*upcloud.ManagedDatabase), nil
}

func buildManagedDatabaseSubResourceID(serviceID, subResourceID string) string {
	return fmt.Sprintf("%s/%s", serviceID, subResourceID)
}

func diffSuppressCreateOnlyProperty(_, _, _ string, d *schema.ResourceData) bool {
	return d.Id() != ""
}

func splitManagedDatabaseSubResourceID(packed string) (serviceID string, subResourceID string) {
	parts := strings.SplitN(packed, "/", 2)
	serviceID = parts[0]
	if len(parts) > 1 {
		subResourceID = parts[1]
	}
	return serviceID, subResourceID
}

func isManagedDatabaseFullyCreated(db *upcloud.ManagedDatabase) bool {
	return db.State == upcloud.ManagedDatabaseStateRunning && len(db.Backups) > 0 && len(db.Users) > 0
}
