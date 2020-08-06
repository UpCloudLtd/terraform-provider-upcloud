package upcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceUpCloudZone() *schema.Resource {
	return &schema.Resource{
		Read: resourceUpCloudZoneRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceUpCloudZoneRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}
