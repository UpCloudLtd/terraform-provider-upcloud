package upcloud

import (
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
)

func checkLogin(c *service.Service) error {
	_, err := c.GetAccount()
	if err != nil {
		serviceError, ok := err.(*upcloud.Error)
		if ok {
			return fmt.Errorf("Error %s: '%s'", serviceError.ErrorCode, serviceError.ErrorMessage)
		}
		return fmt.Errorf("Unspecified error")
	}
	return nil
}
