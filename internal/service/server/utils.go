package server

import (
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
)

func isProviderAccountSubaccount(s *service.Service) (bool, error) {
	account, err := s.GetAccount()
	if err != nil {
		return false, err
	}
	a, err := s.GetAccountDetails(&request.GetAccountDetailsRequest{Username: account.UserName})
	if err != nil {
		return false, err
	}
	return a.IsSubaccount(), nil
}

func serverDefaultTitleFromHostname(hostname string) string {
	const suffix string = " (managed by terraform)"
	if len(hostname)+len(suffix) > serverTitleLength {
		hostname = fmt.Sprintf("%s…", hostname[:serverTitleLength-len(suffix)-1])
	}
	return fmt.Sprintf("%s%s", hostname, suffix)
}
