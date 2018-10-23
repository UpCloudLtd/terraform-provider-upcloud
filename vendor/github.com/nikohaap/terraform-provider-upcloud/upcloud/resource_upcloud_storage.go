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
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudStorage() *schema.Resource {
	return &schema.Resource{
		Create: resourceUpCloudStorageCreate,
		Read:   resourceUpCloudStorageRead,
		Update: resourceUpCloudStorageUpdate,
		Delete: resourceUpCloudStorageDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"access": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"backup_rule": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"interval": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
					"time": &schema.Schema{
						Type:     schema.TypeString,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"retention": &schema.Schema{
						Type:     schema.TypeString,
						Optional: true,
						Default:  false,
					},
				},
			},
			"size": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validateStorageSize,
			},
			"tier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Required: true,
			},
			"license": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceUpCloudStorageCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)

	createStorageRequest := request.CreateStorageRequest{}
	if size, ok := d.GetOk("size"); ok {
		createStorageRequest.Size = size.(int)
	}
	if tier, ok := d.GetOk("tier"); ok {
		createStorageRequest.Tier = tier.(string)
	}
	if title, ok := d.GetOk("title"); ok {
		createStorageRequest.Title = title.(string)
	}
	if zone, ok := d.GetOk("zone"); ok {
		createStorageRequest.Zone = zone.(string)
	}
	if br, ok := d.GetOk("backup_rule"); ok {
		br := br.(map[string]interface{})
		backupRule := upcloud.BackupRule{
			Interval:  br["interval"].(string),
			Time:      br["time"].(string),
			Retention: br["retention"].(int),
		}

		createStorageRequest.BackupRule = &backupRule
	}

	storage, err := client.CreateStorage(&createStorageRequest)

	if err != nil {
		return err
	}

	d.SetId(storage.UUID)

	return resourceUpCloudStorageRead(d, meta)
}

func resourceUpCloudStorageRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	r := &request.GetStorageDetailsRequest{
		UUID: d.Id(),
	}
	storage, err := client.GetStorageDetails(r)

	if err != nil {
		return err
	}

	d.Set("size", storage.Size)
	d.Set("title", storage.Title)
	d.Set("tier", storage.Tier)
	d.Set("zone", storage.Zone)
	d.Set("backup_rule", upcloud.BackupRule{
		Interval:  storage.BackupRule.Interval,
		Time:      storage.BackupRule.Time,
		Retention: storage.BackupRule.Retention,
	})

	return nil
}

func resourceUpCloudStorageUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	r := &request.ModifyStorageRequest{}

	if d.HasChange("size") {
		_, newSize := d.GetChange("size")
		r.Size = newSize.(int)
	}

	if d.HasChange("title") {
		_, newTitle := d.GetChange("title")
		r.Title = newTitle.(string)
	}

	_, err := client.ModifyStorage(r)

	if err != nil {
		return err
	}

	return nil
}

func resourceUpCloudStorageDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)

	deleteStorageRequest := &request.DeleteStorageRequest{
		UUID: d.Id(),
	}
	err := client.DeleteStorage(deleteStorageRequest)

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
	return resource.Retry(time.Minute*15, func() *resource.RetryError {
		r := &request.GetStorageDetailsRequest{
			UUID: UUID,
		}
		storage, err := client.GetStorageDetails(r)
		log.Printf("Waiting for storage %s to come online ...", storage.UUID)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Get server details error: %s", err))
		}

		if storage.State == "error" {
			return resource.NonRetryableError(fmt.Errorf("Storage on the error state: %s", err))
		}

		if storage.State != "online" {
			return resource.RetryableError(fmt.Errorf("Expected storaeg to be created but was in state %s", storage.State))
		}
		return nil
	})
}

func modifyStorage(d *schema.ResourceData, meta interface{}, storageDevice map[string]interface{}) error {
	client := meta.(*service.Service)
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
	return nil
}

func storageUpdateAllowed(meta interface{}, cloneStorageDeviceID string) error {
	//Allowed use cases from public to private or from private to private (Restricted by the UPC API)
	client := meta.(*service.Service)
	newStorageDevice, err := client.GetStorageDetails(
		&request.GetStorageDetailsRequest{
			UUID: cloneStorageDeviceID,
		},
	)

	if err != nil {
		return err
	}

	if newStorageDevice.Access != "private" {
		return errors.New("template update allowed only from public to private or from private to private access storages")
	}

	return nil
}
