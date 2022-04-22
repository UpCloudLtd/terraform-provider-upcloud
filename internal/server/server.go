package server

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/storage"
)

func BuildServerOpts(d *schema.ResourceData, meta interface{}) (*request.CreateServerRequest, error) {
	r := &request.CreateServerRequest{
		Zone:     d.Get("zone").(string),
		Hostname: d.Get("hostname").(string),
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
	if attr, ok := d.GetOk("simple_backup"); ok {
		simpleBackupAttrs := attr.(*schema.Set).List()[0].(map[string]interface{})
		r.SimpleBackup = BuildSimpleBackupOpts(simpleBackupAttrs)
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

func BuildSimpleBackupOpts(attrs map[string]interface{}) string {
	if time, ok := attrs["time"]; ok {
		if plan, ok := attrs["plan"]; ok {
			return fmt.Sprintf("%s,%s", time, plan)
		}
	}

	return "no"
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

func getTagsAsMap(service *service.Service) (map[string]upcloud.Tag, error) {
	currTags := make(map[string]upcloud.Tag)

	resp, err := service.GetTags()
	if err != nil {
		return nil, err
	}

	for _, tag := range resp.Tags {
		// Convert key to lowercase as tag names are case-insensitive
		currTags[strings.ToLower(tag.Name)] = tag
	}

	return currTags, nil
}

// RemoveServerTags removes tags from server and
// deletes the tag if it is not used by any other server.
func RemoveServerTags(service *service.Service, serverUUID string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	if _, err := service.UntagServer(&request.UntagServerRequest{
		UUID: serverUUID,
		Tags: tags,
	}); err != nil {
		return err
	}

	currTags, err := getTagsAsMap(service)
	if err != nil {
		return err
	}

	for _, tagName := range tags {
		// Find tag to be removed
		if tag, ok := currTags[strings.ToLower(tagName)]; ok {
			// Delete tag if it is not used by any servers
			if len(tag.Servers) == 0 {
				if err := service.DeleteTag(&request.DeleteTagRequest{
					Name: tag.Name,
				}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type TagsExistsWarning struct {
	tags map[string]string
}

func (w TagsExistsWarning) Error() string {
	var duplicates []string
	for k, v := range w.tags {
		duplicates = append(duplicates, fmt.Sprintf("%s = %s", k, v))
	}
	return fmt.Sprintf("some tags already exist with different letter casing, existing names will be used (%s)", strings.Join(duplicates, ", "))
}

// AddServerTags creates tags that do not yet exist and tags server with given tags.
func AddServerTags(service *service.Service, serverUUID string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	// Store current tags in map to easily check if a tag exists
	currTags, err := getTagsAsMap(service)
	if err != nil {
		return err
	}

	duplicates := make(map[string]string)

	// Create tags that do not yet exist
	for _, tag := range tags {
		existingTag, ok := currTags[strings.ToLower(tag)]
		if !ok {
			if _, err := service.CreateTag(&request.CreateTagRequest{
				Tag: upcloud.Tag{
					Name: tag,
				},
			}); err != nil {
				return err
			}
		} else if tag != existingTag.Name {
			duplicates[tag] = existingTag.Name
		}
	}

	if _, err := service.TagServer(&request.TagServerRequest{
		UUID: serverUUID,
		Tags: tags,
	}); err != nil {
		return err
	}

	if len(duplicates) != 0 {
		return TagsExistsWarning{duplicates}
	}

	return nil
}

func sliceToMap(input []string) map[string]bool {
	output := make(map[string]bool)
	for _, i := range input {
		output[i] = true
	}
	return output
}

// GetTagChange determines tags to add and delete based on given old and new tags.
// Order of the tags is ignored.
func GetTagChange(oldTags, newTags []string) (tagsToAdd, tagsToDelete []string) {
	olgTagsMap := sliceToMap(oldTags)
	newTagsMap := sliceToMap(newTags)

	for tag := range newTagsMap {
		if _, ok := olgTagsMap[tag]; !ok {
			tagsToAdd = append(tagsToAdd, tag)
		}
	}

	for tag := range olgTagsMap {
		if _, ok := newTagsMap[tag]; !ok {
			tagsToDelete = append(tagsToDelete, tag)
		}
	}

	return
}

func UpdateServerTags(service *service.Service, serverUUID string, oldTags, newTags []string) error {
	tagsToAdd, tagsToDelete := GetTagChange(oldTags, newTags)

	if err := RemoveServerTags(service, serverUUID, tagsToDelete); err != nil {
		return err
	}

	return AddServerTags(service, serverUUID, tagsToAdd)
}
