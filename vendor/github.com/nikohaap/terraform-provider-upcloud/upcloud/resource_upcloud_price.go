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
				Elem: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						ForceNew: true,
					},
					"firewall": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"io_request_backup": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"io_request_hdd": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"io_request_maxiops": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"ipv4_address": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"ipv6_address": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"public_ipv4_bandwidth_in": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"public_ipv4_bandwidth_out": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"public_ipv6_bandwidth_in": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"public_ipv6_bandwidth_out": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"server_core": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"server_memory": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"storage_backup": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"storage_hdd": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"storage_maxiops": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"server_plan_1xCPU-1GB": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
					},
					"server_plan_2xCPU-2GB": {
						Type:     schema.TypeMap,
						Elem:     resourceUpCloudPriceZone(),
						ForceNew: true,
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
