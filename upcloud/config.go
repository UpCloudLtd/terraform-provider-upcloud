package upcloud

import (
	"fmt"
	"log"

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
	res, err := svc.GetAccount()
	if err != nil {
		if serviceError, ok := err.(*upcloud.Error); ok {
			errMsg := fmt.Errorf("Error creating Service object. Error code: %s, Error message: %s",
				serviceError.ErrorCode, serviceError.ErrorMessage)
			return nil, errMsg
		}
	}

	log.Printf("[INFO] UpCloud Client configured for user: %s", res.UserName)
	return svc, nil
}
