package upcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudZone() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"amount": {
				Type: schema.TypeInt,
			},
			"price": {
				Type: schema.TypeFloat,
			},
		},
	}
}
