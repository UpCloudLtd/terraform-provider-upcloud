package upcloud

import (
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
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
				Type:     schema.TypeInt,
				Required: true,
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
