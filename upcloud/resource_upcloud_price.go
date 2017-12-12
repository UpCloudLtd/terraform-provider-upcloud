package upcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudPrice() *schema.Resource {
	return &schema.Resource{
		Read: resourceUpCloudPriceRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"zone": {
				Type: schema.TypeList,
				Elem: map[string]*schema.Schema{
					"name": {
						Type: schema.TypeString,
					},
					"firewall": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"io_request_backup": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"io_request_hdd": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"io_request_maxiops": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"ipv4_address": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"ipv6_address": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"public_ipv4_bandwidth_in": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"public_ipv4_bandwidth_out": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"public_ipv6_bandwidth_in": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"public_ipv6_bandwidth_out": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"server_core": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"server_memory": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"storage_backup": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"storage_hdd": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"storage_maxiops": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"server_plan_1xCPU-1GB": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
					"server_plan_2xCPU-2GB": {
						Type: schema.TypeMap,
						Elem: resourceUpCloudZone(),
					},
				},
			},
		},
	}
}

func resourceUpCloudPriceRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}
