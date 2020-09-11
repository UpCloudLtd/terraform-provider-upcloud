package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
)

// GetStorages returns all available storages
func (s *Service) GetStorages(r *request.GetStoragesRequest) (*upcloud.Storages, error) {
	storages := upcloud.Storages{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	json.Unmarshal(response, &storages)

	return &storages, nil
}

// GetStorageDetails returns extended details about the specified piece of storage
func (s *Service) GetStorageDetails(r *request.GetStorageDetailsRequest) (*upcloud.StorageDetails, error) {
	storageDetails := upcloud.StorageDetails{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	json.Unmarshal(response, &storageDetails)

	return &storageDetails, nil
}

// CreateStorage creates the specified storage
func (s *Service) CreateStorage(r *request.CreateStorageRequest) (*upcloud.StorageDetails, error) {
	storageDetails := upcloud.StorageDetails{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	json.Unmarshal(response, &storageDetails)

	return &storageDetails, nil
}

// ModifyStorage modifies the specified storage device
func (s *Service) ModifyStorage(r *request.ModifyStorageRequest) (*upcloud.StorageDetails, error) {
	storageDetails := upcloud.StorageDetails{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPutRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	json.Unmarshal(response, &storageDetails)

	return &storageDetails, nil
}

// AttachStorage attaches the specified storage to the specified server
func (s *Service) AttachStorage(r *request.AttachStorageRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	json.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// DetachStorage detaches the specified storage from the specified server
func (s *Service) DetachStorage(r *request.DetachStorageRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	json.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// DeleteStorage deletes the specified storage device
func (s *Service) DeleteStorage(r *request.DeleteStorageRequest) error {
	err := s.client.PerformJSONDeleteRequest(s.client.CreateRequestURL(r.RequestURL()))

	if err != nil {
		return parseJSONServiceError(err)
	}

	return nil
}

// CloneStorage detaches the specified storage from the specified server
func (s *Service) CloneStorage(r *request.CloneStorageRequest) (*upcloud.StorageDetails, error) {
	storageDetails := upcloud.StorageDetails{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	json.Unmarshal(response, &storageDetails)

	return &storageDetails, nil
}

// TemplatizeStorage detaches the specified storage from the specified server
func (s *Service) TemplatizeStorage(r *request.TemplatizeStorageRequest) (*upcloud.StorageDetails, error) {
	storageDetails := upcloud.StorageDetails{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	json.Unmarshal(response, &storageDetails)

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
			return nil, fmt.Errorf("timeout reached while waiting for storage to enter state \"%s\"", r.DesiredState)
		}
	}
}

// LoadCDROM loads a storage as a CD-ROM in the CD-ROM device of a server
func (s *Service) LoadCDROM(r *request.LoadCDROMRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	json.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// EjectCDROM ejects the storage from the CD-ROM device of a server
func (s *Service) EjectCDROM(r *request.EjectCDROMRequest) (*upcloud.ServerDetails, error) {
	serverDetails := upcloud.ServerDetails{}
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), nil)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	json.Unmarshal(response, &serverDetails)

	return &serverDetails, nil
}

// CreateBackup creates a backup of the specified storage
func (s *Service) CreateBackup(r *request.CreateBackupRequest) (*upcloud.StorageDetails, error) {
	storageDetails := upcloud.StorageDetails{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	json.Unmarshal(response, &storageDetails)

	return &storageDetails, nil
}

// RestoreBackup creates a backup of the specified storage
func (s *Service) RestoreBackup(r *request.RestoreBackupRequest) error {
	_, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), nil)

	if err != nil {
		return parseJSONServiceError(err)
	}

	return nil
}

// CreateStorageImport begins the process of importing an image onto a storage device. A `upcloud.StorageImportSourceHTTPImport` source
// will import from an HTTP source. `upcloud.StorageImportSourceDirectUpload` will directly upload the file specified in `SourceLocation`.
func (s *Service) CreateStorageImport(r *request.CreateStorageImportRequest) (*upcloud.StorageImportDetails, error) {

	if r.Source == request.StorageImportSourceDirectUpload {
		return s.directStorageImport(r)
	}

	return s.doCreateStorageImport(r)
}

// doCreateStorageImport will POST the CreateStorageImport request and handle the error and normal response.
func (s *Service) doCreateStorageImport(r *request.CreateStorageImportRequest) (*upcloud.StorageImportDetails, error) {
	storageImport := upcloud.StorageImportDetails{}
	requestBody, _ := json.Marshal(r)
	response, err := s.client.PerformJSONPostRequest(s.client.CreateRequestURL(r.RequestURL()), requestBody)

	if err != nil {
		return nil, parseJSONServiceError(err)
	}

	json.Unmarshal(response, &storageImport)

	return &storageImport, nil
}

// directStorageImport handles the direct upload logic including getting the upload URL and PUT the file data
// to that endpoint.
func (s *Service) directStorageImport(r *request.CreateStorageImportRequest) (*upcloud.StorageImportDetails, error) {
	if r.SourceLocation == "" {
		return nil, errors.New("SourceLocation must be specified")
	}

	f, err := os.Open(r.SourceLocation)
	if err != nil {
		return nil, fmt.Errorf("unable to open SourceLocation: %w", err)
	}
	defer f.Close()

	r.SourceLocation = ""
	storageImport, err := s.doCreateStorageImport(r)
	if err != nil {
		return nil, err
	}

	if storageImport.DirectUploadURL == "" {
		return nil, errors.New("no DirectUploadURL found in response")
	}

	_, err = s.client.PerformJSONPutUploadRequest(storageImport.DirectUploadURL, f)
	if err != nil {
		return nil, err
	}

	storageImport, err = s.GetStorageImportDetails(&request.GetStorageImportDetailsRequest{
		UUID: r.StorageUUID,
	})
	if err != nil {
		return nil, err
	}

	return storageImport, nil
}

// GetStorageImportDetails gets updated details about the specified storage import.
func (s *Service) GetStorageImportDetails(r *request.GetStorageImportDetailsRequest) (*upcloud.StorageImportDetails, error) {
	storageDetails := upcloud.StorageImportDetails{}
	response, err := s.basicGetRequest(r.RequestURL())

	if err != nil {
		return nil, err
	}

	json.Unmarshal(response, &storageDetails)

	return &storageDetails, nil
}

// WaitForStorageImportCompletion waits for the importing storage to complete.
func (s *Service) WaitForStorageImportCompletion(r *request.WaitForStorageImportCompletionRequest) (*upcloud.StorageImportDetails, error) {
	attempts := 0
	sleepDuration := time.Second * 5

	for {
		attempts++

		storageImportDetails, err := s.GetStorageImportDetails(&request.GetStorageImportDetailsRequest{
			UUID: r.StorageUUID,
		})

		if err != nil {
			return nil, err
		}

		switch storageImportDetails.State {
		case upcloud.StorageImportStateCompleted:
			return storageImportDetails, nil
		case upcloud.StorageImportStateCancelled,
			upcloud.StorageImportStateCancelling,
			upcloud.StorageImportStateFailed:
			if storageImportDetails.ErrorCode != "" || storageImportDetails.ErrorMessage != "" {
				return storageImportDetails, &upcloud.Error{
					ErrorCode:    storageImportDetails.ErrorCode,
					ErrorMessage: storageImportDetails.ErrorMessage,
				}
			}
			return storageImportDetails, &upcloud.Error{
				ErrorCode:    storageImportDetails.State,
				ErrorMessage: "Storage Import Failed",
			}
		default:
			if time.Duration(attempts)*sleepDuration >= r.Timeout {
				return nil, errors.New("timeout reached while waiting for import to complete")
			}

			time.Sleep(sleepDuration)
		}
	}
}
