package upcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudPlan() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Read:   resourceUpCloudPlanRead,
		Delete: resourceUpCloudPlanDelete,
		Schema: map[string]*schema.Schema{
			"core_number": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},
			"memory_amount": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"public_traffic_out": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},
			"storage_size": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},
			"storage_tier": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}
func resourceUpCloudPlanRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudPlanDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
