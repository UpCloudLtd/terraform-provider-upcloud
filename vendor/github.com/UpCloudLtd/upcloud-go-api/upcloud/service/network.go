package service

import (
	"encoding/json"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
)

type Network interface {
	GetNetworks() (*upcloud.Networks, error)
	GetNetworksInZone(r *request.GetNetworksInZoneRequest) (*upcloud.Networks, error)
	CreateNetwork(r *request.CreateNetworkRequest) (*upcloud.Network, error)
	GetNetworkDetails(r *request.GetNetworkDetailsRequest) (*upcloud.Network, error)
	ModifyNetwork(r *request.ModifyNetworkRequest) (*upcloud.Network, error)
	DeleteNetwork(r *request.DeleteNetworkRequest) error
	GetServerNetworks(r *request.GetServerNetworksRequest) (*upcloud.Networking, error)
	CreateNetworkInterface(r *request.CreateNetworkInterfaceRequest) (*upcloud.Interface, error)
	ModifyNetworkInterface(r *request.ModifyNetworkInterfaceRequest) (*upcloud.Interface, error)
	DeleteNetworkInterface(r *request.DeleteNetworkInterfaceRequest) error
	GetRouters() (*upcloud.Routers, error)
	GetRouterDetails(r *request.GetRouterDetailsRequest) (*upcloud.Router, error)
	CreateRouter(r *request.CreateRouterRequest) (*upcloud.Router, error)
	ModifyRouter(r *request.ModifyRouterRequest) (*upcloud.Router, error)
	DeleteRouter(r *request.DeleteRouterRequest) error
}

var _ Network = (*Service)(nil)

// GetNetworks returns the all the available networks
func (s *Service) GetNetworks() (*upcloud.Networks, error) {
	networks := upcloud.Networks{}
	response, err := s.basicGetRequest("/network")

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &networks)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &networks, nil
}

// GetNetworksInZone returns the all the available networks within the specified zone.
func (s *Service) GetNetworksInZone(r *request.GetNetworksInZoneRequest) (*upcloud.Networks, error) {
	networks := upcloud.Networks{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &networks)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &networks, nil
}

// CreateNetwork creates a new network and returns the network details for the new network.
func (s *Service) CreateNetwork(r *request.CreateNetworkRequest) (*upcloud.Network, error) {
	network := upcloud.Network{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	err = json.Unmarshal(response, &network)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &network, nil
}

// GetNetworkDetails returns the details for the specified network.
func (s *Service) GetNetworkDetails(r *request.GetNetworkDetailsRequest) (*upcloud.Network, error) {
	network := upcloud.Network{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &network)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &network, nil
}

// ModifyNetwork modifies the existing specified network.
func (s *Service) ModifyNetwork(r *request.ModifyNetworkRequest) (*upcloud.Network, error) {
	network := upcloud.Network{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPutRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	err = json.Unmarshal(response, &network)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &network, nil
}

// DeleteNetwork deletes the specified network.
func (s *Service) DeleteNetwork(r *request.DeleteNetworkRequest) error {
	err := s.client.PerformJSONDeleteRequest(s.client.CreateRequestURL(r.RequestURL()))

	if err != nil {
		return parseJSONServiceError(err)
	}

	return nil
}

// GetServerNetworks returns all the networks associated with the specified server.
func (s *Service) GetServerNetworks(r *request.GetServerNetworksRequest) (*upcloud.Networking, error) {
	networking := upcloud.Networking{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &networking)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &networking, nil
}

// CreateNetworkInterface creates a new network interface on the specified server.
func (s *Service) CreateNetworkInterface(r *request.CreateNetworkInterfaceRequest) (*upcloud.Interface, error) {
	iface := upcloud.Interface{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	err = json.Unmarshal(response, &iface)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &iface, nil
}

// ModifyNetworkInterface modifies the specified network interface on the specified server.
func (s *Service) ModifyNetworkInterface(r *request.ModifyNetworkInterfaceRequest) (*upcloud.Interface, error) {
	iface := upcloud.Interface{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPutRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	err = json.Unmarshal(response, &iface)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &iface, nil
}

// DeleteNetworkInterface removes the specified network interface from the specified server.
func (s *Service) DeleteNetworkInterface(r *request.DeleteNetworkInterfaceRequest) error {
	err := s.client.PerformJSONDeleteRequest(s.client.CreateRequestURL(r.RequestURL()))

	if err != nil {
		return parseJSONServiceError(err)
	}

	return nil
}

// GetRouters returns the all the available routers
func (s *Service) GetRouters() (*upcloud.Routers, error) {
	routers := upcloud.Routers{}
	response, err := s.basicGetRequest("/router")

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &routers)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &routers, nil
}

// GetRouterDetails returns the details for the specified router.
func (s *Service) GetRouterDetails(r *request.GetRouterDetailsRequest) (*upcloud.Router, error) {
	router := upcloud.Router{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &router)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &router, nil
}

// CreateRouter creates a new router.
func (s *Service) CreateRouter(r *request.CreateRouterRequest) (*upcloud.Router, error) {
	router := upcloud.Router{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	err = json.Unmarshal(response, &router)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &router, nil
}

// ModifyRouter modifies the configuration of the specified existing router.
func (s *Service) ModifyRouter(r *request.ModifyRouterRequest) (*upcloud.Router, error) {
	router := upcloud.Router{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPatchRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	err = json.Unmarshal(response, &router)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &router, nil
}

// DeleteRouter deletes the specified router.
func (s *Service) DeleteRouter(r *request.DeleteRouterRequest) error {
	err := s.client.PerformJSONDeleteRequest(s.client.CreateRequestURL(r.RequestURL()))

	if err != nil {
		return parseJSONServiceError(err)
	}

	return nil
}
