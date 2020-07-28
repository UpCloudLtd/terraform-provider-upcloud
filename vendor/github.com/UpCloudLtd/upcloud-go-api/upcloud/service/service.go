package service

import (
	"encoding/xml"
	"fmt"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"time"
)

// Service represents the API service. The specified client is used to communicate with the API
type Service struct {
	client *client.Client
}

// New constructs and returns a new service object configured with the specified client
func New(client *client.Client) *Service {
	service := Service{}
	service.client = client

	return &service
}

// GetAccount returns the current user's account
func (s *Service) GetAccount() (*upcloud.Account, error) {
	account := upcloud.Account{}
	response, err := s.basicGetRequest("/account")

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &account)

	return &account, nil
}

// GetZones returns the available zones
func (s *Service) GetZones() (*upcloud.Zones, error) {
	zones := upcloud.Zones{}
	response, err := s.basicGetRequest("/zone")

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &zones)

	return &zones, nil
}

// GetPriceZones returns the available price zones and their corresponding prices
func (s *Service) GetPriceZones() (*upcloud.PrizeZones, error) {
	zones := upcloud.PrizeZones{}
	response, err := s.basicGetRequest("/price")

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &zones)

	return &zones, nil
}

// GetTimeZones returns the available timezones
func (s *Service) GetTimeZones() (*upcloud.TimeZones, error) {
	zones := upcloud.TimeZones{}
	response, err := s.basicGetRequest("/timezone")

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &zones)

	return &zones, nil
}

// GetPlans returns the available service plans
func (s *Service) GetPlans() (*upcloud.Plans, error) {
	plans := upcloud.Plans{}
	response, err := s.basicGetRequest("/plan")

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &plans)

	return &plans, nil
}

// GetServerConfigurations returns the available pre-configured server configurations
func (s *Service) GetServerConfigurations() (*upcloud.ServerConfigurations, error) {
	serverConfigurations := upcloud.ServerConfigurations{}
	response, err := s.basicGetRequest("/server_size")

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverConfigurations)

	return &serverConfigurations, nil
}

// GetServers returns the available servers
func (s *Service) GetServers() (*upcloud.Servers, error) {
	servers := upcloud.Servers{}
	response, err := s.basicGetRequest("/server")

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &servers)

	return &servers, nil
}

// GetServerDetails returns extended details about the specified server
func (s *Service) GetServerDetails(r *request.GetServerDetailsRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// CreateServer creates a server and returns the server details for the newly created server
func (s *Service) CreateServer(r *request.CreateServerRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// WaitForServerState blocks execution until the specified server has entered the specified state. If the state changes
// favorably, the new server details are returned. The method will give up after the specified timeout
func (s *Service) WaitForServerState(r *request.WaitForServerStateRequest) (*upcloud.ServerDetails, error) {
	attempts := 0
	sleepDuration := time.Second * 5

	for {
		// Always wait for one attempt period before querying the state the first time. Newly created servers
		// may not immediately switch to "maintenance" upon creation, triggering a false positive from this
		// method
		attempts++
		time.Sleep(sleepDuration)

		serverDetails, err := s.GetServerDetails(&request.GetServerDetailsRequest{
			UUID: r.UUID,
		})

		if err != nil {
			return nil, err
		}

		// Either wait for the server to enter the desired state or wait for it to leave the undesired state
		if r.DesiredState != "" && serverDetails.State == r.DesiredState {
			return serverDetails, nil
		} else if r.UndesiredState != "" && serverDetails.State != r.UndesiredState {
			return serverDetails, nil
		}

		if time.Duration(attempts)*sleepDuration >= r.Timeout {
			return nil, fmt.Errorf("Timeout reached while waiting for server to enter state \"%s\"", r.DesiredState)
		}
	}
}

// StartServer starts the specified server
func (s *Service) StartServer(r *request.StartServerRequest) (*upcloud.ServerDetails, error) {
	// Increase the client timeout to match the request timeout
	s.client.SetTimeout(r.Timeout)

	serverDetails := upcloud.ServerDetails{}
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), nil)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// StopServer stops the specified server
func (s *Service) StopServer(r *request.StopServerRequest) (*upcloud.ServerDetails, error) {
	// Increase the client timeout to match the request timeout
	s.client.SetTimeout(r.Timeout)

	serverDetails := upcloud.ServerDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// RestartServer restarts the specified server
func (s *Service) RestartServer(r *request.RestartServerRequest) (*upcloud.ServerDetails, error) {
	// Increase the client timeout to match the request timeout
	s.client.SetTimeout(r.Timeout)

	serverDetails := upcloud.ServerDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// ModifyServer modifies the configuration of an existing server. Attaching and detaching storages as well as assigning
// and releasing IP addresses have their own separate operations.
func (s *Service) ModifyServer(r *request.ModifyServerRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPutRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// DeleteServer deletes the specified server
func (s *Service) DeleteServer(r *request.DeleteServerRequest) error {
	err := s.client.PerformDeleteRequest(s.client.CreateRequestUrl(r.RequestURL()))

	if err != nil {
		return parseServiceError(err)
	}

	return nil
}

// TagServer tags a server with with one or more tags
func (s *Service) TagServer(r *request.TagServerRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), nil)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// UntagServer removes one or more tags from a server
func (s *Service) UntagServer(r *request.UntagServerRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), nil)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// CreateTag creates a new tag, optionally assigning it to one or more servers at the same time
func (s *Service) CreateTag(r *request.CreateTagRequest) (*upcloud.Tag, error) {
	tagDetails := upcloud.Tag{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &tagDetails)

	return &tagDetails, nil
}

// ModifyTag modifies a tag (e.g. renaming it)
func (s *Service) ModifyTag(r *request.ModifyTagRequest) (*upcloud.Tag, error) {
	tagDetails := upcloud.Tag{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPutRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &tagDetails)

	return &tagDetails, nil
}

// DeleteTag deletes the specified tag
func (s *Service) DeleteTag(r *request.DeleteTagRequest) error {
	err := s.client.PerformDeleteRequest(s.client.CreateRequestUrl(r.RequestURL()))

	if err != nil {
		return parseServiceError(err)
	}

	return nil
}

// GetStorages returns all available storages
func (s *Service) GetStorages(r *request.GetStoragesRequest) (*upcloud.Storages, error) {
	storages := upcloud.Storages{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	xml.Unmarshal(response, &storages)

	return &storages, nil
}

// GetStorageDetails returns extended details about the specified piece of storage
func (s *Service) GetStorageDetails(r *request.GetStorageDetailsRequest) (*upcloud.StorageDetails, error) {
	storageDetails := upcloud.StorageDetails{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	xml.Unmarshal(response, &storageDetails)

	return &storageDetails, nil
}

// CreateStorage creates the specified storage
func (s *Service) CreateStorage(r *request.CreateStorageRequest) (*upcloud.StorageDetails, error) {
	storageDetails := upcloud.StorageDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &storageDetails)

	return &storageDetails, nil
}

// ModifyStorage modifies the specified storage device
func (s *Service) ModifyStorage(r *request.ModifyStorageRequest) (*upcloud.StorageDetails, error) {
	storageDetails := upcloud.StorageDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPutRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &storageDetails)

	return &storageDetails, nil
}

// AttachStorage attaches the specified storage to the specified server
func (s *Service) AttachStorage(r *request.AttachStorageRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// DetachStorage detaches the specified storage from the specified server
func (s *Service) DetachStorage(r *request.DetachStorageRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// DeleteStorage deletes the specified storage device
func (s *Service) DeleteStorage(r *request.DeleteStorageRequest) error {
	err := s.client.PerformDeleteRequest(s.client.CreateRequestUrl(r.RequestURL()))

	if err != nil {
		return parseServiceError(err)
	}

	return nil
}

// CloneStorage detaches the specified storage from the specified server
func (s *Service) CloneStorage(r *request.CloneStorageRequest) (*upcloud.StorageDetails, error) {
	storageDetails := upcloud.StorageDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &storageDetails)

	return &storageDetails, nil
}

// TemplatizeStorage detaches the specified storage from the specified server
func (s *Service) TemplatizeStorage(r *request.TemplatizeStorageRequest) (*upcloud.StorageDetails, error) {
	storageDetails := upcloud.StorageDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &storageDetails)

	return &storageDetails, nil
}

// WaitForStorageState blocks execution until the specified storage device has entered the specified state. If the
// state changes favorably, the new storage details is returned. The method will give up after the specified timeout
func (s *Service) WaitForStorageState(r *request.WaitForStorageStateRequest) (*upcloud.StorageDetails, error) {
	attempts := 0
	sleepDuration := time.Second * 5

	for {
		attempts++

		storageDetails, err := s.GetStorageDetails(&request.GetStorageDetailsRequest{
			UUID: r.UUID,
		})

		if err != nil {
			return nil, err
		}

		if storageDetails.State == r.DesiredState {
			return storageDetails, nil
		}

		time.Sleep(sleepDuration)

		if time.Duration(attempts)*sleepDuration >= r.Timeout {
			return nil, fmt.Errorf("Timeout reached while waiting for storage to enter state \"%s\"", r.DesiredState)
		}
	}
}

// LoadCDROM loads a storage as a CD-ROM in the CD-ROM device of a server
func (s *Service) LoadCDROM(r *request.LoadCDROMRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// EjectCDROM ejects the storage from the CD-ROM device of a server
func (s *Service) EjectCDROM(r *request.EjectCDROMRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// CreateBackup creates a backup of the specified storage
func (s *Service) CreateBackup(r *request.CreateBackupRequest) (*upcloud.StorageDetails, error) {
	storageDetails := upcloud.StorageDetails{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &storageDetails)

	return &storageDetails, nil
}

// RestoreBackup creates a backup of the specified storage
func (s *Service) RestoreBackup(r *request.RestoreBackupRequest) error {
	requestBody, _ := xml.Marshal(r)
	_, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return parseServiceError(err)
	}

	return nil
}

// GetIPAddresses returns all IP addresses associated with the account
func (s *Service) GetIPAddresses() (*upcloud.IPAddresses, error) {
	ipAddresses := upcloud.IPAddresses{}
	response, err := s.basicGetRequest("/ip_address")

	if err != nil {
		return nil, err
	}

	xml.Unmarshal(response, &ipAddresses)

	return &ipAddresses, nil
}

// GetIPAddressDetails returns extended details about the specified IP address
func (s *Service) GetIPAddressDetails(r *request.GetIPAddressDetailsRequest) (*upcloud.IPAddress, error) {
	ipAddress := upcloud.IPAddress{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	xml.Unmarshal(response, &ipAddress)

	return &ipAddress, nil
}

// AssignIPAddress assigns the specified IP address to the specified server
func (s *Service) AssignIPAddress(r *request.AssignIPAddressRequest) (*upcloud.IPAddress, error) {
	ipAddress := upcloud.IPAddress{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &ipAddress)

	return &ipAddress, nil
}

// ModifyIPAddress modifies the specified IP address
func (s *Service) ModifyIPAddress(r *request.ModifyIPAddressRequest) (*upcloud.IPAddress, error) {
	ipAddress := upcloud.IPAddress{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPutRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &ipAddress)

	return &ipAddress, nil
}

// ReleaseIPAddress releases the specified IP address from the server it is attached to
func (s *Service) ReleaseIPAddress(r *request.ReleaseIPAddressRequest) error {
	err := s.client.PerformDeleteRequest(s.client.CreateRequestUrl(r.RequestURL()))

	if err != nil {
		return parseServiceError(err)
	}

	return nil
}

// GetFirewallRules returns the firewall rules for the specified server
func (s *Service) GetFirewallRules(r *request.GetFirewallRulesRequest) (*upcloud.FirewallRules, error) {
	firewallRules := upcloud.FirewallRules{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	xml.Unmarshal(response, &firewallRules)

	return &firewallRules, nil
}

// GetFirewallRuleDetails returns extended details about the specified firewall rule
func (s *Service) GetFirewallRuleDetails(r *request.GetFirewallRuleDetailsRequest) (*upcloud.FirewallRule, error) {
	firewallRule := upcloud.FirewallRule{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &firewallRule)

	return &firewallRule, nil
}

// CreateFirewallRule creates the firewall rule
func (s *Service) CreateFirewallRule(r *request.CreateFirewallRuleRequest) (*upcloud.FirewallRule, error) {
	firewallRule := upcloud.FirewallRule{}
	requestBody, _ := xml.Marshal(r)
	response, err := s.client.PerformPostRequest(s.client.CreateRequestUrl(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseServiceError(err)
	}

	xml.Unmarshal(response, &firewallRule)

	return &firewallRule, nil
}

// DeleteFirewallRule deletes the specified firewall rule
func (s *Service) DeleteFirewallRule(r *request.DeleteFirewallRuleRequest) error {
	err := s.client.PerformDeleteRequest(s.client.CreateRequestUrl(r.RequestURL()))

	if err != nil {
		return parseServiceError(err)
	}

	return nil
}

// GetTags returns all tags
func (s *Service) GetTags() (*upcloud.Tags, error) {
	tags := upcloud.Tags{}
	response, err := s.basicGetRequest("/tag")

	if err != nil {
		return nil, err
	}

	xml.Unmarshal(response, &tags)

	return &tags, nil
}

// Wrapper that performs a GET request to the specified location and returns the response or a service error
func (s *Service) basicGetRequest(location string) ([]byte, error) {
	requestUrl := s.client.CreateRequestUrl(location)
	response, err := s.client.PerformGetRequest(requestUrl)

	if err != nil {
		return nil, parseServiceError(err)
	}

	return response, nil
}

// Parses an error returned from the client into a service error object
func parseServiceError(err error) error {
	// Parse service errors
	if clientError, ok := err.(*client.Error); ok {
		serviceError := upcloud.Error{}
		responseBody := clientError.ResponseBody
		xml.Unmarshal(responseBody, &serviceError)

		return &serviceError
	}

	return err
}
