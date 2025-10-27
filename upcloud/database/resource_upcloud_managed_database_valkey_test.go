package database

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUpcloudManagedDatabaseValkeyProperties(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/valkey_properties_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/valkey_properties_s2.tf")

	name := "upcloud_managed_database_valkey.valkey_properties"
	prop := func(name string) string {
		return fmt.Sprintf("properties.0.%s", name)
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "plan", "1x1xCPU-2GB"),
					resource.TestCheckResourceAttr(name, "zone", "fi-hel2"),
					resource.TestCheckResourceAttr(name, prop("public_access"), "false"),
					resource.TestCheckResourceAttr(name, prop("valkey_lfu_decay_time"), "2"),
					resource.TestCheckResourceAttr(name, prop("valkey_number_of_databases"), "2"),
					resource.TestCheckResourceAttr(name, prop("valkey_notify_keyspace_events"), "KEA"),
					resource.TestCheckResourceAttr(name, prop("valkey_pubsub_client_output_buffer_limit"), "128"),
					resource.TestCheckResourceAttr(name, prop("valkey_ssl"), "false"),
					resource.TestCheckResourceAttr(name, prop("valkey_lfu_log_factor"), "11"),
					resource.TestCheckResourceAttr(name, prop("valkey_io_threads"), "1"),
					resource.TestCheckResourceAttr(name, prop("valkey_maxmemory_policy"), "allkeys-lru"),
					resource.TestCheckResourceAttr(name, prop("valkey_persistence"), "off"),
					resource.TestCheckResourceAttr(name, prop("valkey_timeout"), "310"),
					resource.TestCheckResourceAttr(name, prop("valkey_acl_channels_default"), "allchannels"),
					resource.TestCheckResourceAttr(name, prop("automatic_utility_network_ip_filter"), "false"),
					resource.TestCheckResourceAttr(name, prop("service_log"), "true"),
				),
			},
			{
				Config: testDataS2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, prop("public_access"), "true"),
					resource.TestCheckResourceAttr(name, prop("valkey_lfu_decay_time"), "1"),
					resource.TestCheckResourceAttr(name, prop("valkey_number_of_databases"), "3"),
					resource.TestCheckResourceAttr(name, prop("valkey_notify_keyspace_events"), ""),
					resource.TestCheckResourceAttr(name, prop("valkey_pubsub_client_output_buffer_limit"), "256"),
					resource.TestCheckResourceAttr(name, prop("valkey_ssl"), "true"),
					resource.TestCheckResourceAttr(name, prop("valkey_lfu_log_factor"), "12"),
					resource.TestCheckResourceAttr(name, prop("valkey_io_threads"), "1"),
					resource.TestCheckResourceAttr(name, prop("valkey_maxmemory_policy"), "volatile-lru"),
					resource.TestCheckResourceAttr(name, prop("valkey_persistence"), "rdb"),
					resource.TestCheckResourceAttr(name, prop("valkey_timeout"), "320"),
					resource.TestCheckResourceAttr(name, prop("valkey_acl_channels_default"), "resetchannels"),
					resource.TestCheckResourceAttr(name, prop("automatic_utility_network_ip_filter"), "true"),
				),
			},
		},
	})
}
