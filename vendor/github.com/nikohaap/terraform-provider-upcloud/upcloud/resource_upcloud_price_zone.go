package upcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudPriceZone() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"amount": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},
			"price": {
				Type:     schema.TypeFloat,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}
