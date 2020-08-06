package upcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceUpCloudTimezone() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Read: resourceUpCloudTimezoneRead,
		Schema: map[string]*schema.Schema{
			"timezone": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func resourceUpCloudTimezoneRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}
