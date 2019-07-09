package upcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudPrice() *schema.Resource {
	return &schema.Resource{
		Read:   resourceUpCloudPriceRead,
		Delete: resourceUpCloudPriceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"zone": {
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"firewall": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"io_request_backup": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"io_request_hdd": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"io_request_maxiops": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"ipv4_address": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"ipv6_address": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"public_ipv4_bandwidth_in": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"public_ipv4_bandwidth_out": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"public_ipv6_bandwidth_in": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"public_ipv6_bandwidth_out": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"server_core": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"server_memory": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"storage_backup": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"storage_hdd": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"storage_maxiops": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"server_plan_1xcpu_1gb": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
						"server_plan_2xcpu_2gb": {
							Type:     schema.TypeMap,
							Elem:     resourceUpCloudPriceZone(),
							ForceNew: true,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceUpCloudPriceRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudPriceDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
