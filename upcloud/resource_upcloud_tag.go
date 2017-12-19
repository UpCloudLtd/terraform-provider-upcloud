package upcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceUpCloudTagCreate,
		Read:   resourceUpCloudTagRead,
		Update: resourceUpCloudTagUpdate,
		Delete: resourceUpCloudTagDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"servers": {
				Type: schema.TypeMap,
				Elem: map[string]*schema.Schema{
					"server": {
						Type:     schema.TypeList,
						Required: true,
						Elem:     resourceUpCloudServer(),
					},
				},
			},
		},
	}
}

func resourceUpCloudTagCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudTagRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudTagUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudTagDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
