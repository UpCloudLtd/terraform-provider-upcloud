package upcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudPlan() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Read: resourceUpCloudPlanRead,
		Schema: map[string]*schema.Schema{
			"core_number": {
				Type: schema.TypeInt,
			},
			"memory_amount": {
				Type: schema.TypeInt,
			},
			"name": {
				Type: schema.TypeString,
			},
			"public_traffic_out": {
				Type: schema.TypeInt,
			},
			"storage_size": {
				Type: schema.TypeInt,
			},
			"storage_tier": {
				Type: schema.TypeString,
			},
		},
	}
}
func resourceUpCloudPlanRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}
