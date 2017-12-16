package upcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceUpCloudIPCreate,
		Read:   resourceUpCloudIPRead,
		Update: resourceUpCloudIPUpdate,
		Delete: resourceUpCloudIPDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"access": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ptr_record": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"part_of_plan": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUpCloudIPCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudIPRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudIPUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudIPDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
