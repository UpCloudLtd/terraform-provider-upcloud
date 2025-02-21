package upcloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Config struct {
	Username string
	Password string
	Token    string
}

func (c *Config) Client() (*service.Service, error) {
	client := client.New("", "", utils.WithAuth(c.Username, c.Password, c.Token))
	svc := service.New(client)
	res, err := c.checkLogin(svc)
	if err != nil {
		return nil, err
	}
	tflog.Info(context.Background(), "UpCloud Client configured", map[string]interface{}{"user": res.UserName})
	return svc, nil
}

func (c *Config) checkLogin(svc *service.Service) (*upcloud.Account, error) {
	const numRetries = 10
	var (
		err     error
		res     *upcloud.Account
		problem *upcloud.Problem
	)

	for trys := 0; trys < numRetries; trys++ {
		res, err = svc.GetAccount(context.Background())
		if err == nil {
			break
		}
		if errors.As(err, &problem) && problem.Status == http.StatusUnauthorized {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	if err != nil {
		svcErr, ok := err.(*upcloud.Problem)
		if ok {
			return nil, fmt.Errorf("[ERROR] Failed to get account, error was %s: '%s'", svcErr.ErrorCode(), svcErr.Title)
		}
		return nil, err
	}

	return res, nil
}
