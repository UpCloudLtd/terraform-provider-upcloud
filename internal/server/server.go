package server

import (
	"fmt"
	"log"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/storage"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func BuildServerOpts(d *schema.ResourceData, meta interface{}) (*request.CreateServerRequest, error) {
	r := &request.CreateServerRequest{
		Zone:     d.Get("zone").(string),
		Hostname: d.Get("hostname").(string),
		Title:    fmt.Sprintf("%s (managed by terraform)", d.Get("hostname").(string)),
	}

	if attr, ok := d.GetOk("firewall"); ok {
		if attr.(bool) {
			r.Firewall = "on"
		} else {
			r.Firewall = "off"
		}
	}
	if attr, ok := d.GetOk("metadata"); ok {
		if attr.(bool) {
			r.Metadata = upcloud.True
		} else {
			r.Metadata = upcloud.False
		}
	}
	if attr, ok := d.GetOk("cpu"); ok {
		r.CoreNumber = attr.(int)
	}
	if attr, ok := d.GetOk("mem"); ok {
		r.MemoryAmount = attr.(int)
	}
	if attr, ok := d.GetOk("user_data"); ok {
		r.UserData = attr.(string)
	}
	if attr, ok := d.GetOk("plan"); ok {
		r.Plan = attr.(string)
	}
	if login, ok := d.GetOk("login"); ok {
		loginOpts, deliveryMethod, err := buildLoginOpts(login, meta)
		if err != nil {
			return nil, err
		}
		r.LoginUser = loginOpts
		r.PasswordDelivery = deliveryMethod
	}

	r.Host = d.Get("host").(int)

	if template, ok := d.GetOk("template.0"); ok {
		template := template.(map[string]interface{})
		if template["title"].(string) == "" {
			template["title"] = fmt.Sprintf("terraform-%s-disk", r.Hostname)
		}
		serverStorageDevice := request.CreateServerStorageDevice{
			Action:  "clone",
			Address: template["address"].(string),
			Size:    template["size"].(int),
			Storage: template["storage"].(string),
			Title:   template["title"].(string),
		}
		if attr, ok := d.GetOk("template.0.backup_rule.0"); ok {
			serverStorageDevice.BackupRule = storage.BackupRule(attr.(map[string]interface{}))
		}
		if source := template["storage"].(string); source != "" {
			// Assume template name is given and attempt map name to UUID
			if _, err := uuid.ParseUUID(source); err != nil {
				l, err := meta.(*service.Service).GetStorages(
					&request.GetStoragesRequest{
						Type: upcloud.StorageTypeTemplate,
					})

				if err != nil {
					return nil, err
				}
				for _, s := range l.Storages {
					if s.Title == source {
						source = s.UUID
						break
					}
				}
			}

			serverStorageDevice.Storage = source
		}
		r.StorageDevices = append(r.StorageDevices, serverStorageDevice)
	}

	if storageDevices, ok := d.GetOk("storage_devices"); ok {
		storageDevices := storageDevices.(*schema.Set)
		for _, storageDevice := range storageDevices.List() {
			storageDevice := storageDevice.(map[string]interface{})
			r.StorageDevices = append(r.StorageDevices, request.CreateServerStorageDevice{
				Action:  "attach",
				Address: storageDevice["address"].(string),
				Type:    storageDevice["type"].(string),
				Storage: storageDevice["storage"].(string),
			})
		}
	}

	networking, err := buildNetworkOpts(d, meta)
	if err != nil {
		return nil, err
	}

	r.Networking = &request.CreateServerNetworking{
		Interfaces: networking,
	}

	return r, nil
}

func buildLoginOpts(v interface{}, meta interface{}) (*request.LoginUser, string, error) {
	// Construct LoginUser struct from the schema
	r := &request.LoginUser{}
	e := v.(*schema.Set).List()[0]
	m := e.(map[string]interface{})

	// Set username as is
	r.Username = m["user"].(string)

	// Set 'create_password' to "yes" or "no" depending on the bool value.
	// Would be nice if the API would just get a standard bool str.
	createPassword := "no"
	b := m["create_password"].(bool)
	if b {
		createPassword = "yes"
	}
	r.CreatePassword = createPassword

	// Handle SSH keys one by one
	keys := make([]string, 0)
	for _, k := range m["keys"].([]interface{}) {
		key := k.(string)
		keys = append(keys, key)
	}
	r.SSHKeys = keys

	// Define password delivery method none/email/sms
	deliveryMethod := m["password_delivery"].(string)

	return r, deliveryMethod, nil
}

func VerifyServerStopped(stopRequest request.StopServerRequest, meta interface{}) error {
	if stopRequest.Timeout == 0 {
		stopRequest.Timeout = time.Minute * 2
	}
	if stopRequest.StopType == "" {
		stopRequest.StopType = upcloud.StopTypeSoft
	}

	client := meta.(*service.Service)
	// Get current server state
	r := &request.GetServerDetailsRequest{
		UUID: stopRequest.UUID,
	}
	server, err := client.GetServerDetails(r)
	if err != nil {
		return err
	}
	if server.State != upcloud.ServerStateStopped {
		// Soft stop with 2 minute timeout, after which hard stop occurs
		log.Printf("[INFO] Stopping server (server UUID: %s)", stopRequest.UUID)
		_, err := client.StopServer(&stopRequest)
		if err != nil {
			return err
		}
		_, err = client.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:         stopRequest.UUID,
			DesiredState: upcloud.ServerStateStopped,
			Timeout:      time.Minute * 5,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func VerifyServerStarted(startRequest request.StartServerRequest, meta interface{}) error {
	if startRequest.Timeout == 0 {
		startRequest.Timeout = time.Minute * 2
	}

	client := meta.(*service.Service)
	// Get current server state
	r := &request.GetServerDetailsRequest{
		UUID: startRequest.UUID,
	}
	server, err := client.GetServerDetails(r)
	if err != nil {
		return err
	}
	if server.State != upcloud.ServerStateStarted {
		log.Printf("[INFO] Starting server (server UUID: %s)", startRequest.UUID)
		_, err := client.StartServer(&startRequest)
		if err != nil {
			return err
		}
		_, err = client.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:         startRequest.UUID,
			DesiredState: upcloud.ServerStateStarted,
			Timeout:      time.Minute * 5,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
