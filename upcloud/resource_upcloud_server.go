package upcloud

import (
	"log"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
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
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"title": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUpCloudServerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	if err := checkLogin(client); err != nil {
		return err
	}
	server, err := client.CreateServer(&request.CreateServerRequest{
		Zone:             d.Get("zone").(string),
		Title:            d.Get("title").(string),
		Hostname:         d.Get("hostname").(string),
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
		},
	})
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
	return nil
}

func resourceUpCloudServerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	if err := checkLogin(client); err != nil {
		return err
	}
	return nil
}

func resourceUpCloudServerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	if err := checkLogin(client); err != nil {
		return err
	}
	return nil
}

func resourceUpCloudServerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	if err := checkLogin(client); err != nil {
		return err
	}
	return nil
}
