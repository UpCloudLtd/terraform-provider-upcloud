package upcloud

import (
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
				Required: true,
				ForceNew: true,
			},
			"backup_rule": {
				Type: schema.TypeMap,
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
				Required: true,
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
				Required: true,
			},
		},
	}
}

func resourceUpCloudStorageCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudStorageRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudStorageUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudStorageDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
