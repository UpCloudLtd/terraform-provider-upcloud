package service

import (
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// The service object used by the tests
var svc *Service

// TestMain is the main test method
func TestMain(m *testing.M) {
	setup()
	retCode := m.Run()

	// Optionally perform teardown
	deleteResources := os.Getenv("UPCLOUD_GO_SDK_TEST_DELETE_RESOURCES")
	if deleteResources == "yes" {
		log.Print("UPCLOUD_GO_SDK_TEST_DELETE_RESOURCES defined, deleting all resources ...")
		teardown()
	}

	os.Exit(retCode)
}

// Configures the test environment
func setup() {
	user, password := getCredentials()

	c := client.New(user, password)
	c.SetTimeout(time.Second * 30)
	svc = New(c)
}

// Tears down the test environment by removing all resources
func teardown() {
	log.Print("Deleting all servers ...")
	servers, err := svc.GetServers()
	handleError(err)

	for _, server := range servers.Servers {
		// Try to ensure the server is not in maintenance state
		log.Printf("Waiting for server with UUID %s to leave maintenance state ...", server.UUID)
		serverDetails, err := svc.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:           server.UUID,
			UndesiredState: upcloud.ServerStateMaintenance,
			Timeout:        time.Minute * 5,
		})
		handleError(err)

		// Stop the server if it's still running
		if serverDetails.State != upcloud.ServerStateStopped {
			log.Printf("Stopping server with UUID %s ...", server.UUID)
			stopServer(server.UUID)
		}

		// Delete the server
		log.Printf("Deleting the server with UUID %s ...", server.UUID)
		deleteServer(server.UUID)
	}

	// Delete all private storage devices
	log.Print("Deleting all storage devices ...")
	storages, err := svc.GetStorages(&request.GetStoragesRequest{
		Access: upcloud.StorageAccessPrivate,
	})
	handleError(err)

	for _, storage := range storages.Storages {
		// Wait for the storage to come online so we can delete it
		if storage.State != upcloud.StorageStateOnline {
			log.Printf("Waiting for storage %s to come online ...", storage.UUID)
			_, err = svc.WaitForStorageState(&request.WaitForStorageStateRequest{
				UUID:         storage.UUID,
				DesiredState: upcloud.StorageStateOnline,
				Timeout:      time.Minute * 5,
			})
			handleError(err)
		}

		log.Printf("Deleting the storage with UUID %s ...", storage.UUID)
		deleteStorage(storage.UUID)
	}

	// Delete all tags
	log.Print("Deleting all tags ...")
	deleteAllTags()
}

// TestGetAccount tests that the GetAccount() method returns proper data
func TestGetAccount(t *testing.T) {
	account, err := svc.GetAccount()
	username, _ := getCredentials()
	handleError(err)

	if account.UserName != username {
		t.Errorf("TestGetAccount expected %s, got %s", username, account.UserName)
	}
}

// TestErrorHandling checks that the correct error type is returned from service methods
func TestErrorHandling(t *testing.T) {
	// Perform a bogus request that will certainly fail
	_, err := svc.StartServer(&request.StartServerRequest{
		UUID: "invalid",
	})

	// Check that the correct error type is returned
	expectedErrorType := "*upcloud.Error"
	actualErrorType := reflect.TypeOf(err).String()

	if actualErrorType != expectedErrorType {
		t.Errorf("TestErrorHandling expected %s, got %s", expectedErrorType, actualErrorType)
	}
}

// TestCreateModifyDeleteServer performs the following actions:
//
// - creates a server
// - modifies the server
// - stops the server
// - deletes the server
func TestCreateModifyDeleteServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	// Create a server
	serverDetails := createServer("TestCreateModifyDeleteServer")
	t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

	// Modify the server
	t.Log("Modifying the server ...")

	serverDetails, err := svc.ModifyServer(&request.ModifyServerRequest{
		UUID:  serverDetails.UUID,
		Title: "Modified server",
	})

	handleError(err)
	t.Log("Waiting for the server to exit maintenance state ...")

	serverDetails, err = svc.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStarted,
		Timeout:      time.Minute * 5,
	})

	handleError(err)
	t.Logf("Server is now modified, new title is %s", serverDetails.Title)

	// Stop the server
	t.Logf("Stopping server with UUID %s ...", serverDetails.UUID)
	stopServer(serverDetails.UUID)
	t.Log("Server is now stopped")

	// Delete the server
	t.Logf("Deleting the server with UUID %s...", serverDetails.UUID)
	deleteServer(serverDetails.UUID)
	t.Log("Server is now deleted")
}

// TestCreateModifyDeleteStorage performs the following actions:
//
// - creates a new storage disk
// - modifies the storage
// - deletes the storage
func TestCreateModifyDeleteStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	// Create some storage
	storageDetails := createStorage()
	t.Logf("Storage %s with UUID %s created", storageDetails.Title, storageDetails.UUID)

	// Modify the storage
	t.Log("Modifying the storage ...")

	storageDetails, err := svc.ModifyStorage(&request.ModifyStorageRequest{
		UUID:  storageDetails.UUID,
		Title: "New fancy title",
	})

	handleError(err)
	t.Logf("Storage with UUID %s modified successfully, new title is %s", storageDetails.UUID, storageDetails.Title)

	// Delete the storage
	t.Log("Deleting the storage ...")
	deleteStorage(storageDetails.UUID)
	t.Log("Storage is now deleted")
}

// TestAttachDetachStorage performs the following actions:
//
// - creates a server
// - stops the server
// - creates a new storage disk
// - attaches the storage
// - detaches the storage
// - deletes the storage
// - deletes the server
func TestAttachDetachStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	// Create a server
	serverDetails := createServer("TestAttachDetachStorage")
	t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

	// Stop the server
	t.Logf("Stopping server with UUID %s ...", serverDetails.UUID)
	stopServer(serverDetails.UUID)
	t.Log("Server is now stopped")

	// Create some storage
	storageDetails := createStorage()
	t.Logf("Storage %s with UUID %s created", storageDetails.Title, storageDetails.UUID)

	// Attach the storage
	t.Logf("Attaching storage %s", storageDetails.UUID)

	serverDetails, err := svc.AttachStorage(&request.AttachStorageRequest{
		StorageUUID: storageDetails.UUID,
		ServerUUID:  serverDetails.UUID,
		Type:        upcloud.StorageTypeDisk,
		Address:     "scsi:0:0",
	})

	handleError(err)
	t.Logf("Storage attached to server with UUID %s", serverDetails.UUID)

	// Detach the storage
	t.Logf("Detaching storage %s", storageDetails.UUID)

	serverDetails, err = svc.DetachStorage(&request.DetachStorageRequest{
		ServerUUID: serverDetails.UUID,
		Address:    "scsi:0:0",
	})

	handleError(err)
	t.Logf("Storage %s detached", storageDetails.UUID)
}

// TestCloneStorage performs the following actions:
//
// - creates a storage device
// - clones the storage device
// - deletes the clone and the storage device
func TestCloneStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	// Create storage
	storageDetails := createStorage()
	t.Logf("Storage %s with UUID %s created", storageDetails.Title, storageDetails.UUID)

	// Clone the storage
	t.Log("Cloning storage ...")

	clonedStorageDetails, err := svc.CloneStorage(&request.CloneStorageRequest{
		UUID:  storageDetails.UUID,
		Title: "Cloned storage",
		Zone:  "fi-hel1",
		Tier:  upcloud.StorageTierMaxIOPS,
	})

	handleError(err)
	waitForStorageOnline(clonedStorageDetails.UUID)
	t.Logf("Storage cloned as %s", clonedStorageDetails.UUID)
}

// TestTemplatizeServerStorage performs the following actions:
//
// - creates a server
// - templatizes the server's storage
// - deletes the new storage
// - stops and deletes the server
func TestTemplatizeServerStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	// Create server
	serverDetails := createServer("TestTemplatizeServerStorage")
	t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

	// Stop the server
	t.Logf("Stopping server with UUID %s ...", serverDetails.UUID)
	stopServer(serverDetails.UUID)
	t.Log("Server is now stopped")

	// Get extended service details
	serverDetails, err := svc.GetServerDetails(&request.GetServerDetailsRequest{
		UUID: serverDetails.UUID,
	})

	handleError(err)

	// Templatize the server's first storage device
	storageFound := false
	for i, storage := range serverDetails.StorageDevices {
		if i == 0 {
			storageFound = true
			t.Log("Templatizing storage ...")

			storageDetails, err := svc.TemplatizeStorage(&request.TemplatizeStorageRequest{
				UUID:  storage.UUID,
				Title: "Templatized storage",
			})

			handleError(err)
			waitForStorageOnline(storageDetails.UUID)
			t.Logf("Storage templatized as %s", storageDetails.UUID)

			break
		}
	}

	// Fail the test if for some reason the storage was never found
	if !storageFound {
		t.FailNow()
	}
}

// TestLoadEjectCDROM performs the following actions:
//
// - creates a server
// - stops the server
// - attaches a CD-ROM device
// - loads a CD-ROM
// - ejects the CD-ROM
func TestLoadEjectCDROM(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	// Create the server
	serverDetails := createServer("TestLoadEjectCDROM")
	t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

	// Stop the server
	t.Logf("Stopping server with UUID %s ...", serverDetails.UUID)
	stopServer(serverDetails.UUID)
	t.Log("Server is now stopped")

	// Attach CD-ROM device
	t.Logf("Attaching CD-ROM device to server with UUID %s", serverDetails.UUID)
	serverDetails, err := svc.AttachStorage(&request.AttachStorageRequest{
		ServerUUID: serverDetails.UUID,
		Type:       upcloud.StorageTypeCDROM,
	})

	handleError(err)
	t.Log("CD-ROM is now attached")

	// Load the CD-ROM
	t.Log("Loading CD-ROM into CD-ROM device")
	serverDetails, err = svc.LoadCDROM(&request.LoadCDROMRequest{
		ServerUUID:  serverDetails.UUID,
		StorageUUID: "01000000-0000-4000-8000-000030060101",
	})

	handleError(err)
	t.Log("CD-ROM is now loaded")

	// Eject the CD-ROM
	t.Log("Ejecting CD-ROM from CD-ROM device")
	serverDetails, err = svc.EjectCDROM(&request.EjectCDROMRequest{
		ServerUUID: serverDetails.UUID,
	})

	handleError(err)
	t.Log("CD-ROM is now ejected")
}

// TestCreateRestoreBackup performs the following actions:
//
// - creates a storage device
// - creates a backup of the storage device
//
// It's not feasible to test backup restoration due to time constraints
func TestCreateBackup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	// Create the storage
	storageDetails := createStorage()
	t.Logf("Storage %s with UUID %s created", storageDetails.Title, storageDetails.UUID)

	// Create a backup
	t.Logf("Creating backup of storage with UUID %s ...", storageDetails.UUID)
	backupDetails, err := svc.CreateBackup(&request.CreateBackupRequest{
		UUID:  storageDetails.UUID,
		Title: "Backup",
	})

	handleError(err)
	waitForStorageOnline(storageDetails.UUID)
	t.Logf("Created backup with UUID %s", backupDetails.UUID)
}

// TestAttachModifyReleaseIPAddress performs the following actions
//
// - creates a server
// - assigns an additional IP address to it
// - modifies the PTR record of the IP address
// - deletes the IP address
func TestAttachModifyReleaseIPAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	// Create the server
	serverDetails := createServer("TestAttachModifyReleaseIPAddress")
	t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

	// Stop the server
	t.Logf("Stopping server with UUID %s ...", serverDetails.UUID)
	stopServer(serverDetails.UUID)
	t.Log("Server is now stopped")

	// Assign an IP address
	t.Log("Assigning IP address to server ...")
	ipAddress, err := svc.AssignIPAddress(&request.AssignIPAddressRequest{
		Access:     upcloud.IPAddressAccessPublic,
		Family:     upcloud.IPAddressFamilyIPv6,
		ServerUUID: serverDetails.UUID,
	})
	handleError(err)
	t.Logf("Assigned IP address %s to server with UUID %s", ipAddress.Address, serverDetails.UUID)

	// Modify the PTR record
	t.Logf("Modifying PTR record for address %s ...", ipAddress.Address)
	ipAddress, err = svc.ModifyIPAddress(&request.ModifyIPAddressRequest{
		IPAddress: ipAddress.Address,
		PTRRecord: "such.pointer.example.com",
	})
	handleError(err)
	t.Logf("PTR record modified, new record is %s", ipAddress.PTRRecord)

	// Release the IP address
	t.Log("Releasing the IP address ...")
	err = svc.ReleaseIPAddress(&request.ReleaseIPAddressRequest{
		IPAddress: ipAddress.Address,
	})
	handleError(err)
	t.Log("The IP address is now released")
}

// TestFirewallRules performs the following actions:
//
// - creates a server
// - adds a firewall rule to the server
// - gets details about the firewall rule
// - deletes the firewall rule
//
func TestFirewallRules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	// Create the server
	serverDetails := createServer("TestFirewallRules")
	t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

	// Create firewall rule
	t.Logf("Creating firewall rule #1 for server with UUID %s ...", serverDetails.UUID)
	firewallRule, err := svc.CreateFirewallRule(&request.CreateFirewallRuleRequest{
		ServerUUID: serverDetails.UUID,
		FirewallRule: upcloud.FirewallRule{
			Direction: upcloud.FirewallRuleDirectionIn,
			Action:    upcloud.FirewallRuleActionAccept,
			Family:    upcloud.IPAddressFamilyIPv4,
			Protocol:  upcloud.FirewallRuleProtocolTCP,
			Position:  1,
			Comment:   "This is the comment",
		},
	})
	handleError(err)
	t.Log("Firewall rule created")

	// Get details about the rule
	t.Log("Getting details about firewall rule #1 ...")
	firewallRule, err = svc.GetFirewallRuleDetails(&request.GetFirewallRuleDetailsRequest{
		ServerUUID: serverDetails.UUID,
		Position:   1,
	})
	handleError(err)
	t.Logf("Got firewall rule details, comment is %s", firewallRule.Comment)

	// Delete the firewall rule
	t.Log("Deleting firewall rule #1 ...")
	err = svc.DeleteFirewallRule(&request.DeleteFirewallRuleRequest{
		ServerUUID: serverDetails.UUID,
		Position:   1,
	})
	handleError(err)
	t.Log("Firewall rule #1 deleted")
}

// TestTagging tests that all tagging-related functionality works correctly. It performs the following actions:
//   - creates a server
//   - creates three tags
//   - assigns the first tag to the server
//   - renames the second tag
//   - deletes the third tag
//   - untags the first tag from the server
func TestTagging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	// Create the server
	serverDetails := createServer("TestTagging")
	t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

	// Remove all existing tags
	t.Log("Deleting any existing tags ...")
	deleteAllTags()

	// Create three tags
	tags := []string{
		"tag1",
		"tag2",
		"tag3",
	}

	for _, tag := range tags {
		t.Logf("Creating tag %s", tag)
		tagDetails, err := svc.CreateTag(&request.CreateTagRequest{
			Tag: upcloud.Tag{
				Name: tag,
			},
		})

		handleError(err)
		t.Logf("Tag %s created", tagDetails.Name)
	}

	// Assign the first tag to the server
	serverDetails, err := svc.TagServer(&request.TagServerRequest{
		UUID: serverDetails.UUID,
		Tags: []string{
			"tag1",
		},
	})

	handleError(err)
	t.Logf("Server %s is now tagged with tag %s", serverDetails.Title, "tag1")

	// Rename the second tag
	tagDetails, err := svc.ModifyTag(&request.ModifyTagRequest{
		Name: "tag2",
		Tag: upcloud.Tag{
			Name: "tag2_renamed",
		},
	})

	handleError(err)
	t.Logf("Tag tag2 renamed to %s", tagDetails.Name)

	// Delete the third tag
	err = svc.DeleteTag(&request.DeleteTagRequest{
		Name: "tag3",
	})

	handleError(err)
	t.Log("Tag tag3 deleted")

	// Untag the server
	t.Logf("Removing tag %s from server %s", "tag1", serverDetails.UUID)
	serverDetails, err = svc.UntagServer(&request.UntagServerRequest{
		UUID: serverDetails.UUID,
		Tags: []string{
			"tag1",
		},
	})

	handleError(err)
	t.Logf("Server %s is now untagged", serverDetails.Title)
}

// Creates a server and returns the details about it, panic if creation fails
func createServer(name string) *upcloud.ServerDetails {
	title := "uploud-go-sdk-integration-test-" + name
	hostname := strings.ToLower(title + ".example.com")

	createServerRequest := request.CreateServerRequest{
		Zone:             "fi-hel1",
		Title:            title,
		Hostname:         hostname,
		PasswordDelivery: request.PasswordDeliveryNone,
		StorageDevices: []upcloud.CreateServerStorageDevice{
			{
				Action:  upcloud.CreateServerStorageDeviceActionClone,
				Storage: "01000000-0000-4000-8000-000030060200",
				Title:   "disk1",
				Size:    30,
				Tier:    upcloud.StorageTierMaxIOPS,
			},
		},
		IPAddresses: []request.CreateServerIPAddress{
			{
				Access: upcloud.IPAddressAccessPrivate,
				Family: upcloud.IPAddressFamilyIPv4,
			},
			{
				Access: upcloud.IPAddressAccessPublic,
				Family: upcloud.IPAddressFamilyIPv4,
			},
			{
				Access: upcloud.IPAddressAccessPublic,
				Family: upcloud.IPAddressFamilyIPv6,
			},
		},
	}

	// Create the server and block until it has started
	serverDetails, err := svc.CreateServer(&createServerRequest)

	if err != nil {
		panic(err)
	}

	// Wait for the server to start
	serverDetails, err = svc.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStarted,
		Timeout:      time.Minute * 5,
	})

	handleError(err)

	return serverDetails
}

// Stops the specified server (forcibly)
func stopServer(uuid string) {
	serverDetails, err := svc.StopServer(&request.StopServerRequest{
		UUID:     uuid,
		Timeout:  time.Minute * 5,
		StopType: request.ServerStopTypeHard,
	})

	handleError(err)

	serverDetails, err = svc.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStopped,
		Timeout:      time.Minute * 5,
	})

	handleError(err)
}

// Deletes the specified server
func deleteServer(uuid string) {
	err := svc.DeleteServer(&request.DeleteServerRequest{
		UUID: uuid,
	})

	handleError(err)
}

// Creates a piece of storage and returns the details about it, panic if creation fails
func createStorage() *upcloud.StorageDetails {
	createStorageRequest := request.CreateStorageRequest{
		Tier:  upcloud.StorageTierMaxIOPS,
		Title: "Test storage",
		Size:  10,
		Zone:  "fi-hel1",
		BackupRule: &upcloud.BackupRule{
			Interval:  upcloud.BackupRuleIntervalDaily,
			Time:      "0430",
			Retention: 30,
		},
	}

	storageDetails, err := svc.CreateStorage(&createStorageRequest)

	if err != nil {
		panic(err)
	}

	return storageDetails
}

// Deletes the specified storage
func deleteStorage(uuid string) {
	err := svc.DeleteStorage(&request.DeleteStorageRequest{
		UUID: uuid,
	})

	handleError(err)
}

// deleteAllTags deletes all existing tags
func deleteAllTags() {
	tags, err := svc.GetTags()
	handleError(err)

	for _, tagDetails := range tags.Tags {
		err = svc.DeleteTag(&request.DeleteTagRequest{
			Name: tagDetails.Name,
		})

		handleError(err)
	}
}

// Waits for the specified storage to come online
func waitForStorageOnline(uuid string) {
	_, err := svc.WaitForStorageState(&request.WaitForStorageStateRequest{
		UUID:         uuid,
		DesiredState: upcloud.StorageStateOnline,
		Timeout:      time.Minute * 5,
	})

	handleError(err)
}

// Handles the error by panicing, thus stopping the test execution
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

// Reads the API username and password from the environment, panics if they are not available
func getCredentials() (string, string) {
	user := os.Getenv("UPCLOUD_GO_SDK_TEST_USER")
	password := os.Getenv("UPCLOUD_GO_SDK_TEST_PASSWORD")

	if user == "" || password == "" {
		panic("Unable to retrieve credentials from the environment, ensure UPCLOUD_GO_SDK_TEST_USER and UPCLOUD_GO_SDK_TEST_PASSWORD are exported")
	}

	return user, password
}
