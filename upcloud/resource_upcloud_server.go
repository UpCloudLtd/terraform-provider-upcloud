package upcloud

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceUpCloudServerCreate,
		Read:   resourceUpCloudServerRead,
		Update: resourceUpCloudServerUpdate,
		Delete: resourceUpCloudServerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Update: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"title": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"firewall": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cpu": {
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validateCPUCount,
			},
			"mem": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateStorageSize,
			},
			"template": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"private_networking": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"ipv4": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"ipv6": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"ipv4_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv4_address_private": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"plan": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"storage_devices": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"address": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"action": {
							Type:     schema.TypeString,
							Required: true,
						},
						"size": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validateStorageSize,
						},
						"tier": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"title": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"storage": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"backup_rule": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"interval": {
										Type:     schema.TypeString,
										Required: true,
									},
									"time": {
										Type:     schema.TypeString,
										Required: true,
									},
									"retention": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"login": {
				Type:     schema.TypeSet,
				ForceNew: true,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"keys": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"create_password": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"password_delivery": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "none",
						},
					},
				},
			},
		},
	}
}

func resourceUpCloudServerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	r, err := buildServerOpts(d, meta)
	if err != nil {
		return err
	}
	server, err := client.CreateServer(r)
	if err != nil {
		return err
	}

	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		serverDetails, err := client.GetServerDetails(&request.GetServerDetailsRequest{
			UUID: server.UUID,
		})

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Get server details error: %s", err))
		}

		if serverDetails.State == "error" {
			return resource.NonRetryableError(fmt.Errorf("Instance on the error state: %s", err))
		}

		if serverDetails.State != "started" {
			return resource.RetryableError(fmt.Errorf("Expected instance to be created but was in state %s", serverDetails.State))
		}

		d.SetId(serverDetails.UUID)
		err = buildAfterServerCreationOps(d, client)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error describing instance: %s", err))
		}
		return resource.NonRetryableError(resourceUpCloudServerRead(d, meta))
	})
}

func resourceUpCloudServerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	r := &request.GetServerDetailsRequest{
		UUID: d.Id(),
	}
	server, err := client.GetServerDetails(r)
	if err != nil {
		return err
	}
	d.Set("hostname", server.Hostname)
	d.Set("title", server.Title)
	d.Set("zone", server.Zone)
	d.Set("cpu", server.CoreNumber)
	d.Set("mem", server.MemoryAmount)

	// Store server addresses into state
	for _, ip := range server.IPAddresses {
		if ip.Access == upcloud.IPAddressAccessPrivate && ip.Family == upcloud.IPAddressFamilyIPv4 {
			d.Set("ipv4_address_private", ip.Address)
		}
		if ip.Access == upcloud.IPAddressAccessPublic && ip.Family == upcloud.IPAddressFamilyIPv4 {
			d.Set("ipv4_address", ip.Address)
		}
		if ip.Access == upcloud.IPAddressAccessPublic && ip.Family == upcloud.IPAddressFamilyIPv6 {
			d.Set("ipv6_address", ip.Address)
		}
	}

	storageDevices := d.Get("storage_devices").([]interface{})
	log.Printf("[DEBUG] Configured storage devices in state: %v", storageDevices)
	log.Printf("[DEBUG] Actual storage devices on server: %v", server.StorageDevices)
	for i, storageDevice := range storageDevices {
		storageDevice := storageDevice.(map[string]interface{})
		storageDevice["id"] = server.StorageDevices[i].UUID
		storageDevice["address"] = server.StorageDevices[i].Address
		storageDevice["title"] = server.StorageDevices[i].Title
		if storageDevice["size"] != 0 {
			storageDevice["size"] = server.StorageDevices[i].Size
		}
	}
	d.Set("storage_devices", storageDevices)

	return nil
}

func updateStorageDevices(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	oldStorageDevicesI, storageDevicesI := d.GetChange("storage_devices")
	log.Printf("[DEBUG] JEEEE: %v", oldStorageDevicesI)
	log.Printf("[DEBUG] JOOOO: %v", storageDevicesI)
	d.Set("storage_devices", storageDevicesI)
	storageDevices := storageDevicesI.([]interface{})
	oldStorageDevices := oldStorageDevicesI.([]interface{})
	log.Printf("[DEBUG] New storage devices: %v", storageDevices)
	log.Printf("[DEBUG] Current storage devices: %v", oldStorageDevices)
	for i, storageDevice := range storageDevices {
		storageDevice := storageDevice.(map[string]interface{})
		log.Printf("[DEBUG] Number of current storage devices: %v\n", len(oldStorageDevices))
		var oldStorageDeviceN int
		for i, oldStorageDevice := range oldStorageDevices {
			id1 := oldStorageDevice.(map[string]interface{})["id"].(string)
			id2 := storageDevice["id"].(string)
			log.Printf("[DEBUG] Storage device Id 1: %v, Id 2: %v, Equal: %v", id1, id2, id1 == id2)
			if id1 == id2 {
				oldStorageDeviceN = i
				break
			}
		}

		log.Printf("[DEBUG] Old storage device number: %v\n", oldStorageDeviceN)
		var oldStorageDevice map[string]interface{}
		if oldStorageDeviceN < len(oldStorageDevices) {
			oldStorageDevice = oldStorageDevices[oldStorageDeviceN].(map[string]interface{})
		}
		log.Printf("[DEBUG] New storage device: %v\n", storageDevice)
		log.Printf("[DEBUG] Current storage device: %v\n", oldStorageDevice)
		if oldStorageDevice == nil {
			var newStorageDeviceID string
			switch storageDevice["action"] {
			case upcloud.CreateServerStorageDeviceActionCreate:
				storage, err := buildStorage(storageDevice, i, meta, d.Get("hostname").(string), d.Get("zone").(string))
				if err != nil {
					return err
				}
				newStorage, err := client.CreateStorage(&request.CreateStorageRequest{
					Size:  storage.Size,
					Tier:  storage.Tier,
					Title: storage.Title,
					Zone:  d.Get("zone").(string),
				})
				if err != nil {
					return err
				}
				newStorageDeviceID = newStorage.UUID
				break
			case upcloud.CreateServerStorageDeviceActionClone:
				newStorageDeviceID = storageDevice["storage"].(string)
				break
			case upcloud.CreateServerStorageDeviceActionAttach:
				newStorageDeviceID = storageDevice["storage"].(string)
				break
			}

			attachStorageRequest := request.AttachStorageRequest{
				ServerUUID:  d.Id(),
				StorageUUID: newStorageDeviceID,
				Address:     storageDevice["address"].(string),
			}

			if storageType := storageDevice["type"].(string); storageType != "" {
				attachStorageRequest.Type = storageType
			}

			log.Printf("[DEBUG] Attach storage request: %v", attachStorageRequest)

			client.AttachStorage(&attachStorageRequest)
		} else {
			if canModify, err := canModifyStorage(d, meta, storageDevice["id"].(string)); canModify {
				log.Printf("[DEBUG] Try to modify storage device %v", storageDevice)
				modifyStorage := &request.ModifyStorageRequest{
					UUID:  storageDevice["id"].(string),
					Size:  storageDevice["size"].(int),
					Title: storageDevice["title"].(string),
				}
				if backupRule := storageDevice["backup_rule"].(map[string]interface{}); backupRule != nil && len(backupRule) != 0 {
					log.Println("[DEBUG] Backup rule create")
					retention, err := strconv.Atoi(backupRule["retention"].(string))
					if err != nil {
						return err
					}

					modifyStorage.BackupRule = &upcloud.BackupRule{
						Interval:  backupRule["interval"].(string),
						Retention: retention,
						Time:      backupRule["time"].(string),
					}
				}
				log.Printf("[DEBUG] Storage modify request: %v\n", modifyStorage)
				if _, err := client.ModifyStorage(modifyStorage); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}

			oldStorageDevices = append(oldStorageDevices[:oldStorageDeviceN], oldStorageDevices[oldStorageDeviceN+1:]...)

			if oldStorageDevice["address"] != storageDevice["address"] {
				log.Printf("[DEBUG] Trying to change address from %v to %v", oldStorageDevice["address"], storageDevice["address"])
				client.DetachStorage(&request.DetachStorageRequest{
					ServerUUID: d.Id(),
					Address:    oldStorageDevice["address"].(string),
				})
				client.AttachStorage(&request.AttachStorageRequest{
					ServerUUID:  d.Id(),
					StorageUUID: storageDevice["id"].(string),
					Address:     storageDevice["address"].(string),
				})
			}

			if oldStorageDevice["storage"] != storageDevice["storage"] {
				log.Printf("[DEBUG] Trying to change strorage from %v to %v", oldStorageDevice["storage"], storageDevice["storage"])

				switch storageDevice["action"] {
				case upcloud.CreateServerStorageDeviceActionAttach:
					err := updateStorageAttach(d, meta, i, oldStorageDevice["id"].(string), storageDevice)
					if err != nil {
						return err
					}
				case upcloud.CreateServerStorageDeviceActionClone:
					storageDeviceDetails, err := createStorageClone(d, meta, storageDevice)
					if err != nil {
						return err
					}

					r := &request.DetachStorageRequest{
						ServerUUID: d.Id(),
						Address:    oldStorageDevice["address"].(string),
					}
					if _, err = client.DetachStorage(r); err != nil {
						return err
					}

					if err := updateStorageClone(d, meta, storageDevice, storageDeviceDetails.UUID); err != nil {
						client.AttachStorage(&request.AttachStorageRequest{
							ServerUUID:  d.Id(),
							StorageUUID: storageDevice["id"].(string),
							Address:     storageDevice["address"].(string),
						})
						return err
					}
				}
			}

		}
	}
	log.Printf("[DEBUG] Current storage devices: %v\n", oldStorageDevices)
	for _, oldStorageDevice := range oldStorageDevices {
		oldStorageDevice := oldStorageDevice.(map[string]interface{})
		client.DetachStorage(&request.DetachStorageRequest{
			ServerUUID: d.Id(),
			Address:    oldStorageDevice["address"].(string),
		})
		if oldStorageDevice["action"] != upcloud.CreateServerStorageDeviceActionAttach {
			client.DeleteStorage(&request.DeleteStorageRequest{
				UUID: oldStorageDevice["id"].(string),
			})
		}
	}
	return nil
}

func updateInstanceHarware(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	_, newCPU := d.GetChange("cpu")
	_, newMem := d.GetChange("mem")
	_, newFirewall := d.GetChange("firewall")

	r := &request.ModifyServerRequest{
		UUID: d.Id(),
	}

	if newFirewall.(bool) {
		r.Firewall = "on"
	} else {
		r.Firewall = "off"
	}

	if newCPU != 0 || newMem != 0 {
		log.Printf("[DEBUG] Modifying server, cpu = %v, mem = %v", newCPU, newMem)
		if newCPU != 0 {
			r.CoreNumber = strconv.Itoa(newCPU.(int))
		}
		if newMem != 0 {
			r.MemoryAmount = strconv.Itoa(newMem.(int))
		}
	}
	_, err := client.ModifyServer(r)
	if err != nil {
		return err
	}
	return nil
}

func updateInstanceHarwarePlan(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	_, newPlan := d.GetChange("plan")

	r := &request.ModifyServerRequest{
		UUID: d.Id(),
	}

	r.Plan = newPlan.(string)

	_, err := client.ModifyServer(r)
	if err != nil {
		return err
	}
	return nil
}

func resourceUpCloudServerUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := verifyServerStopped(d, meta); err != nil {
		return err
	}

	if d.HasChange("storage_devices") {
		if err := updateStorageDevices(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("mem") || d.HasChange("cpu") || d.HasChange("firewall") {
		if err := updateInstanceHarware(d, meta); err != nil {
			return err
		}
	}
	if d.HasChange("plan") {
		if err := updateInstanceHarwarePlan(d, meta); err != nil {
			return err
		}
	}

	if err := verifyServerStarted(d, meta); err != nil {
		return err
	}
	return resourceUpCloudServerRead(d, meta)
}

func resourceUpCloudServerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	// Verify server is stopped before deletion
	if err := verifyServerStopped(d, meta); err != nil {
		return err
	}
	// Delete server
	deleteServerRequest := &request.DeleteServerRequest{
		UUID: d.Id(),
	}
	log.Printf("[INFO] Deleting server (server UUID: %s)", d.Id())
	err := client.DeleteServer(deleteServerRequest)
	if err != nil {
		return err
	}

	storageDevices := d.Get("storage_devices").([]interface{})
	for _, storageDevice := range storageDevices {
		// Delete server root disk
		storageDevice := storageDevice.(map[string]interface{})
		id := storageDevice["id"].(string)
		action := storageDevice["action"].(string)
		if action != upcloud.CreateServerStorageDeviceActionAttach {
			deleteStorageRequest := &request.DeleteStorageRequest{
				UUID: id,
			}
			log.Printf("[INFO] Deleting server storage (storage UUID: %s)", id)
			err = client.DeleteStorage(deleteStorageRequest)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func buildServerOpts(d *schema.ResourceData, meta interface{}) (*request.CreateServerRequest, error) {
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

	storageDevices := d.Get("storage_devices").([]interface{})
	storageOpts, err := buildStorageOpts(storageDevices, meta, d.Get("hostname").(string), d.Get("zone").(string))
	if err != nil {
		return nil, err
	}
	r.StorageDevices = storageOpts

	networkOpts, err := buildNetworkOpts(d, meta)
	if err != nil {
		return nil, err
	}
	r.IPAddresses = networkOpts

	return r, nil
}

func serverRestartIsRequired(storageDevices []interface{}) bool {
	for _, storageDevice := range storageDevices {
		storageDevice := storageDevice.(map[string]interface{})
		if backupRule := storageDevice["backup_rule"].(map[string]interface{}); backupRule != nil && len(backupRule) != 0 {
			return true
		}
	}

	return false
}

func buildAfterServerCreationOps(d *schema.ResourceData, meta interface{}) error {
	/*
		Some of the operations such as backup_rule for storage device can only be done after
		the server creation.
	*/

	storageDevices := d.Get("storage_devices").([]interface{})
	if serverRestartIsRequired(storageDevices) {
		if err := verifyServerStopped(d, meta); err != nil {
			return err
		}

		err := buildStorageBackupRuleOps(d, meta)
		if err != nil {
			return err
		}
	}

	if err := verifyServerStarted(d, meta); err != nil {
		return err
	}

	return nil
}

func buildStorageBackupRuleOps(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)

	storageDevices := d.Get("storage_devices").([]interface{})

	r := &request.GetServerDetailsRequest{
		UUID: d.Id(),
	}
	server, err := client.GetServerDetails(r)
	if err != nil {
		return err
	}

	for i, storageDevice := range storageDevices {
		storageDevice := storageDevice.(map[string]interface{})

		if backupRule := storageDevice["backup_rule"].(map[string]interface{}); backupRule != nil && len(backupRule) != 0 {
			retention, err := strconv.Atoi(backupRule["retention"].(string))
			if err != nil {
				return err
			}
			modifyStorage := &request.ModifyStorageRequest{
				UUID: server.StorageDevices[i].UUID,
			}

			modifyStorage.BackupRule = &upcloud.BackupRule{
				Interval:  backupRule["interval"].(string),
				Retention: retention,
				Time:      backupRule["time"].(string),
			}
			client.ModifyStorage(modifyStorage)
		}
	}

	return nil
}

func buildStorage(storageDevice map[string]interface{}, i int, meta interface{}, hostname, zone string) (*upcloud.CreateServerStorageDevice, error) {
	osDisk := upcloud.CreateServerStorageDevice{}

	if source := storageDevice["storage"].(string); source != "" {
		_, err := uuid.ParseUUID(source)
		// Assume template name is given and map name to UUID
		if err != nil {
			client := meta.(*service.Service)
			r := &request.GetStoragesRequest{
				Type: upcloud.StorageTypeTemplate,
			}
			l, err := client.GetStorages(r)
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

		osDisk.Storage = source
	}

	// Set size or use the one defined by target template
	if size := storageDevice["size"].(int); size > 0 {
		osDisk.Size = size
	}

	// Autogenerate disk title
	if title := storageDevice["title"].(string); title != "" {
		osDisk.Title = title
	} else {
		osDisk.Title = fmt.Sprintf("terraform-%s-disk-%d", hostname, i)
	}

	// Set disk tier or use the one defined by target template
	if tier := storageDevice["tier"]; tier != "" {
		osDisk.Tier = tier.(string)
	}

	if storageType := storageDevice["type"].(string); storageType != "" {
		osDisk.Type = storageType
	}

	if address := storageDevice["address"].(string); address != "" {
		osDisk.Address = address
	}

	osDisk.Action = storageDevice["action"].(string)

	log.Printf("[DEBUG] Disk: %v", osDisk)

	return &osDisk, nil
}

func buildStorageOpts(storageDevices []interface{}, meta interface{}, hostname, zone string) ([]upcloud.CreateServerStorageDevice, error) {
	storageCfg := make([]upcloud.CreateServerStorageDevice, 0)
	for i, storageDevice := range storageDevices {
		storageDevice, err := buildStorage(storageDevice.(map[string]interface{}), i, meta, hostname, zone)

		if err != nil {
			return nil, err
		}

		storageCfg = append(storageCfg, *storageDevice)
	}

	return storageCfg, nil
}

func buildNetworkOpts(d *schema.ResourceData, meta interface{}) ([]request.CreateServerIPAddress, error) {
	ifaceCfg := make([]request.CreateServerIPAddress, 0)
	if attr, ok := d.GetOk("ipv4"); ok {
		publicIPv4 := attr.(bool)
		if publicIPv4 {
			publicIPv4 := request.CreateServerIPAddress{
				Access: upcloud.IPAddressAccessPublic,
				Family: upcloud.IPAddressFamilyIPv4,
			}
			ifaceCfg = append(ifaceCfg, publicIPv4)
		}
	}
	if attr, ok := d.GetOk("private_networking"); ok {
		setPrivateIP := attr.(bool)
		if setPrivateIP {
			privateIPv4 := request.CreateServerIPAddress{
				Access: upcloud.IPAddressAccessPrivate,
				Family: upcloud.IPAddressFamilyIPv4,
			}
			ifaceCfg = append(ifaceCfg, privateIPv4)
		}
	}
	if attr, ok := d.GetOk("ipv6"); ok {
		publicIPv6 := attr.(bool)
		if publicIPv6 {
			publicIPv6 := request.CreateServerIPAddress{
				Access: upcloud.IPAddressAccessPublic,
				Family: upcloud.IPAddressFamilyIPv6,
			}
			ifaceCfg = append(ifaceCfg, publicIPv6)
		}
	}
	return ifaceCfg, nil
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

func verifyServerStopped(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)

	// Soft stop with 5 minute timeout, after which hard stop occurs
	return resource.Retry(time.Minute*5, func() *resource.RetryError {
		serverDetails, err := client.GetServerDetails(&request.GetServerDetailsRequest{
			UUID: d.Id(),
		})

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error describing instance: %s", err))
		}

		if serverDetails.State == "error" {
			return resource.NonRetryableError(fmt.Errorf("Instance on the error state: %s", err))
		}

		switch serverDetails.State {
		case "started":
			stopRequest := &request.StopServerRequest{
				UUID:     d.Id(),
				StopType: "soft",
			}
			log.Printf("[INFO] Stopping server (server UUID: %s)", d.Id())
			_, err := client.StopServer(stopRequest)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("Error describing instance: %s", err))
			}
			return resource.RetryableError(fmt.Errorf("Expected instance to be stopped but was in state started"))
		default:
			if serverDetails.State != "stopped" {
				time.Sleep(time.Second * 5)
				return resource.RetryableError(fmt.Errorf("Expected instance to be stopped but was in state %s", serverDetails.State))
			}
		}
		return nil
	})
}

func verifyServerStarted(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)

	return resource.Retry(time.Minute*5, func() *resource.RetryError {
		serverDetails, err := client.GetServerDetails(&request.GetServerDetailsRequest{
			UUID: d.Id(),
		})

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error describing instance: %s", err))
		}

		if serverDetails.State == "error" {
			return resource.NonRetryableError(fmt.Errorf("Instance on the error state: %s", err))
		}

		switch serverDetails.State {
		case "stopped":
			startRequest := &request.StartServerRequest{
				UUID: d.Id(),
			}
			log.Printf("[INFO] Starting server (server UUID: %s)", d.Id())
			_, err := client.StartServer(startRequest)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("Error describing instance: %s", err))
			}
			return resource.RetryableError(fmt.Errorf("Expected instance to be started but was in state stopped"))
		default:
			if serverDetails.State != "started" {
				return resource.RetryableError(fmt.Errorf("Expected instance to be started but was in state %s", serverDetails.State))
			}
		}
		return nil
	})
}

func updateStorageAttach(d *schema.ResourceData, meta interface{}, i int, oldStorageDeviceID string, storageDevice map[string]interface{}) error {
	log.Printf("[DEBUG] ATTACH")

	err1 := errors.New("Attach operation not allowed when updating storage template.")

	return err1
}

func createStorageClone(d *schema.ResourceData, meta interface{}, storageDevice map[string]interface{}) (*upcloud.StorageDetails, error) {
	log.Printf("[DEBUG] CREATE CLONE")

	client := meta.(*service.Service)

	newStorage, err := client.CloneStorage(&request.CloneStorageRequest{
		UUID:  storageDevice["storage"].(string),
		Tier:  storageDevice["tier"].(string),
		Title: storageDevice["title"].(string),
		Zone:  d.Get("zone").(string),
	})

	if err := verifyStorageOnline(d, meta, newStorage.UUID); err != nil {
		return nil, err
	}

	return newStorage, err
}

func updateStorageClone(d *schema.ResourceData, meta interface{}, storageDevice map[string]interface{}, NewStorageDeviceUUID string) error {
	log.Printf("[DEBUG] UPDATE CLONE")

	client := meta.(*service.Service)

	attachStorageRequest := request.AttachStorageRequest{
		ServerUUID:  d.Id(),
		StorageUUID: NewStorageDeviceUUID,
		Address:     storageDevice["address"].(string),
	}

	if storageType := storageDevice["type"].(string); storageType != "" {
		attachStorageRequest.Type = storageType
	}

	if err := verifyStorageOnline(d, meta, NewStorageDeviceUUID); err != nil {
		return err
	}

	log.Printf("[DEBUG] Attach storage request: %v", attachStorageRequest)

	_, err := client.AttachStorage(&attachStorageRequest)

	if err != nil {
		return err
	}

	return nil
}

func canModifyStorage(d *schema.ResourceData, meta interface{}, UUID string) (bool, error) {
	client := meta.(*service.Service)
	r := &request.GetStorageDetailsRequest{
		UUID: UUID,
	}
	storage, err := client.GetStorageDetails(r)
	if err != nil {
		return false, err
	}

	if canModifyAccess := storage.Access; canModifyAccess == "private" {
		return true, nil
	}
	return false, nil
}

func verifyStorageOnline(d *schema.ResourceData, meta interface{}, UUID string) error {
	client := meta.(*service.Service)
	r := &request.GetStorageDetailsRequest{
		UUID: UUID,
	}
	storage, err := client.GetStorageDetails(r)

	if err != nil {
		return err
	}

	if storage.State != upcloud.StorageStateOnline {
		log.Printf("Waiting for storage %s to come online ...", storage.UUID)
		_, err = client.WaitForStorageState(&request.WaitForStorageStateRequest{
			UUID:         storage.UUID,
			DesiredState: upcloud.StorageStateOnline,
			Timeout:      time.Minute * 15,
		})
		return err
	}
	return nil
}
