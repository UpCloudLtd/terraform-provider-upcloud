package service

import (
	"encoding/json"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
)

// GetHosts returns the all the available private hosts
func (s *Service) GetHosts() (*upcloud.Hosts, error) {
	hosts := upcloud.Hosts{}
	response, err := s.basicGetRequest("/host")

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &hosts)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &hosts, nil
}

// GetHostDetails returns the details for a single private host
func (s *Service) GetHostDetails(r *request.GetHostDetailsRequest) (*upcloud.Host, error) {
	host := upcloud.Host{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &host)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &host, nil
}

// ModifyHost modifies the configuration of an existing host.
func (s *Service) ModifyHost(r *request.ModifyHostRequest) (*upcloud.Host, error) {
	host := upcloud.Host{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPatchRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	err = json.Unmarshal(response, &host)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &host, nil
}
