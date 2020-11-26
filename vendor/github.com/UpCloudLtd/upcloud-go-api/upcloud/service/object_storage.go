package service

import (
	"encoding/json"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
)

type ObjectStorage interface {
	GetObjectStorages() (*upcloud.ObjectStorages, error)
	GetObjectStorageDetails(r *request.GetObjectStorageDetailsRequest) (*upcloud.ObjectStorageDetails, error)
	CreateObjectStorage(r *request.CreateObjectStorageRequest) (*upcloud.ObjectStorageDetails, error)
	ModifyObjectStorage(r *request.ModifyObjectStorageRequest) (*upcloud.ObjectStorageDetails, error)
	DeleteObjectStorage(r *request.DeleteObjectStorageRequest) error
}

var _ ObjectStorage = (*Service)(nil)

// GetObjectStorages returns the available objects storages
func (s *Service) GetObjectStorages() (*upcloud.ObjectStorages, error) {
	objectStorages := upcloud.ObjectStorages{}
	response, err := s.basicGetRequest("/object-storage")

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &objectStorages)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &objectStorages, nil
}

// GetObjectStorageDetails returns extended details about the specified Object Storage
func (s *Service) GetObjectStorageDetails(r *request.GetObjectStorageDetailsRequest) (*upcloud.ObjectStorageDetails, error) {
	objectStorageDetails := upcloud.ObjectStorageDetails{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &objectStorageDetails)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &objectStorageDetails, nil
}

// CreateObjectStorage creates a Object Storage and return the Object Storage details for the newly created device
func (s *Service) CreateObjectStorage(r *request.CreateObjectStorageRequest) (*upcloud.ObjectStorageDetails, error) {
	objectStorageDetails := upcloud.ObjectStorageDetails{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	err = json.Unmarshal(response, &objectStorageDetails)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &objectStorageDetails, nil
}

// ModifyObjectStorage modifies the configuration of an existing Object Storage
func (s *Service) ModifyObjectStorage(r *request.ModifyObjectStorageRequest) (*upcloud.ObjectStorageDetails, error) {
	objectStorageDetails := upcloud.ObjectStorageDetails{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPatchRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	json.Unmarshal(response, &objectStorageDetails)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %s, %w", string(response), err)
	}

	return &objectStorageDetails, nil
}

// DeleteObjectStorage deletes the specific Object Storage
func (s *Service) DeleteObjectStorage(r *request.DeleteObjectStorageRequest) error {
	err := s.client.PerformJSONDeleteRequest(s.client.CreateRequestURL(r.RequestURL()))

	if err != nil {
		return parseJSONServiceError(err)
	}

	return nil
}
