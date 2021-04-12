package upcloud

import (
	"fmt"
	"log"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
)

type Config struct {
	Username string
	Password string
}

func (c *Config) Client() (*service.Service, error) {
	client := client.New(c.Username, c.Password)
	svc := service.New(client)
	res, err := c.checkLogin(svc)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] UpCloud Client configured for user: %s", res.UserName)
	return svc, nil
}

func (c *Config) checkLogin(svc *service.Service) (*upcloud.Account, error) {
	const numRetries = 10
	var (
		err error
		res *upcloud.Account
	)

	for trys := 0; trys < numRetries; trys++ {
		res, err = svc.GetAccount()
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	if err != nil {
		svcErr, ok := err.(*upcloud.Error)
		if ok {
			return nil, fmt.Errorf("[ERROR] Failed to get account, error was %s: '%s'", svcErr.ErrorCode, svcErr.ErrorMessage)
		}
		return nil, fmt.Errorf("[ERROR] Failed to get account due to unspecified error")
	}

	return res, nil
}
