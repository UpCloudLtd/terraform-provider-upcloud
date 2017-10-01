package upcloud

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform/helper/schema"
	uuid "github.com/satori/go.uuid"
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
			"cpu": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"mem": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"os_disk_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"os_disk_uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_disk_tier": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "maxiops",
			},
			"template": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
	d.SetId(server.UUID)
	log.Printf("[INFO] Server %s with UUID %s created", server.Title, server.UUID)

	server, err = client.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         server.UUID,
		DesiredState: upcloud.ServerStateStarted,
		Timeout:      time.Minute * 5,
	})
	if err != nil {
		return err
	}
	return resourceUpCloudServerRead(d, meta)
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
	d.Set("region", server.Zone)
	d.Set("cpu", server.CoreNumber)
	d.Set("mem", server.MemoryAmount)

	// TODO: Handle additional disks
	osDisk := server.StorageDevices[0]
	d.Set("os_disk_size", osDisk.Size)
	d.Set("os_disk_uuid", osDisk.UUID)

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
	return nil
}

func resourceUpCloudServerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	if d.HasChange("mem") || d.HasChange("cpu") {
		_, newCPU := d.GetChange("cpu")
		_, newMem := d.GetChange("mem")
		if err := verifyServerStopped(d, meta); err != nil {
			return err
		}
		r := &request.ModifyServerRequest{
			UUID:         d.Id(),
			CoreNumber:   strconv.Itoa(newCPU.(int)),
			MemoryAmount: strconv.Itoa(newMem.(int)),
		}
		_, err := client.ModifyServer(r)
		if err != nil {
			return err
		}
		if err := verifyServerStarted(d, meta); err != nil {
			return err
		}

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
	// Delete server root disk
	rootDiskUUID := d.Get("os_disk_uuid").(string)
	deleteStorageRequest := &request.DeleteStorageRequest{
		UUID: rootDiskUUID,
	}
	log.Printf("[INFO] Deleting server root disk (storage UUID: %s)", rootDiskUUID)
	err = client.DeleteStorage(deleteStorageRequest)
	if err != nil {
		return err
	}
	return nil
}

func buildServerOpts(d *schema.ResourceData, meta interface{}) (*request.CreateServerRequest, error) {
	r := &request.CreateServerRequest{
		Zone:     d.Get("zone").(string),
		Hostname: d.Get("hostname").(string),
		Title:    fmt.Sprintf("%s (managed by terraform)", d.Get("hostname").(string)),
	}

	if attr, ok := d.GetOk("cpu"); ok {
		r.CoreNumber = attr.(int)
	}
	if attr, ok := d.GetOk("mem"); ok {
		r.MemoryAmount = attr.(int)
	}
	if attr, ok := d.GetOk("userdata"); ok {
		r.UserData = attr.(string)
	}

	storageOpts, err := buildStorageOpts(d, meta)
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

func buildStorageOpts(d *schema.ResourceData, meta interface{}) ([]upcloud.CreateServerStorageDevice, error) {
	storageCfg := make([]upcloud.CreateServerStorageDevice, 0)
	source := d.Get("template").(string)
	_, err := uuid.FromString(source)
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
	osDisk := upcloud.CreateServerStorageDevice{
		Action:  upcloud.CreateServerStorageDeviceActionClone,
		Storage: source,
	}

	// Set size or use the one defined by target template
	if attr, ok := d.GetOk("os_disk_size"); ok {
		osDisk.Size = attr.(int)
	}

	// Autogenerate disk title
	osDisk.Title = fmt.Sprintf("terraform-os-disk")

	// Set disk tier or use the one defined by target template
	if attr, ok := d.GetOk("os_disk_tier"); ok {
		tier := attr.(string)
		switch tier {
		case upcloud.StorageTierMaxIOPS:
			osDisk.Tier = upcloud.StorageTierMaxIOPS
		case upcloud.StorageTierHDD:
			osDisk.Tier = upcloud.StorageTierHDD
		default:
			return nil, fmt.Errorf("Invalid disk tier '%s'", tier)
		}
	}
	storageCfg = append(storageCfg, osDisk)

	// TODO: Handle additional disks
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

func verifyServerStopped(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	// Get current server state
	r := &request.GetServerDetailsRequest{
		UUID: d.Id(),
	}
	server, err := client.GetServerDetails(r)
	if err != nil {
		return err
	}
	if server.State != upcloud.ServerStateStopped {
		// Soft stop with 2 minute timeout, after which hard stop occurs
		stopRequest := &request.StopServerRequest{
			UUID:     d.Id(),
			StopType: "soft",
			Timeout:  time.Minute * 2,
		}
		log.Printf("[INFO] Stopping server (server UUID: %s)", d.Id())
		_, err := client.StopServer(stopRequest)
		if err != nil {
			return err
		}
		_, err = client.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:         d.Id(),
			DesiredState: upcloud.ServerStateStopped,
			Timeout:      time.Minute * 5,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func verifyServerStarted(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	// Get current server state
	r := &request.GetServerDetailsRequest{
		UUID: d.Id(),
	}
	server, err := client.GetServerDetails(r)
	if err != nil {
		return err
	}
	if server.State != upcloud.ServerStateStarted {
		// Soft stop with 2 minute timeout, after which hard stop occurs
		startRequest := &request.StartServerRequest{
			UUID:    d.Id(),
			Timeout: time.Minute * 2,
		}
		log.Printf("[INFO] Stopping server (server UUID: %s)", d.Id())
		_, err := client.StartServer(startRequest)
		if err != nil {
			return err
		}
		_, err = client.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:         d.Id(),
			DesiredState: upcloud.ServerStateStarted,
			Timeout:      time.Minute * 5,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
