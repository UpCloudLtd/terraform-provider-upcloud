package upcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudTimezone() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
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
