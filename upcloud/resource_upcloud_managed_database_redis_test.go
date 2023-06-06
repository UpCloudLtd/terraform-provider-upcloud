package upcloud

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudManagedDatabaseRedisProperties(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/redis_properties_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/redis_properties_s2.tf")

	var providers []*schema.Provider
	name := "upcloud_managed_database_redis.redis_properties"
	prop := func(name string) string {
		return fmt.Sprintf("properties.0.%s", name)
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "plan", "2x4xCPU-28GB"),
					resource.TestCheckResourceAttr(name, "zone", "fi-hel2"),
					resource.TestCheckResourceAttr(name, prop("public_access"), "false"),
					resource.TestCheckResourceAttr(name, prop("redis_lfu_decay_time"), "2"),
					resource.TestCheckResourceAttr(name, prop("redis_number_of_databases"), "2"),
					resource.TestCheckResourceAttr(name, prop("redis_notify_keyspace_events"), "KEA"),
					resource.TestCheckResourceAttr(name, prop("redis_pubsub_client_output_buffer_limit"), "128"),
					resource.TestCheckResourceAttr(name, prop("redis_ssl"), "false"),
					resource.TestCheckResourceAttr(name, prop("redis_lfu_log_factor"), "11"),
					resource.TestCheckResourceAttr(name, prop("redis_io_threads"), "2"),
					resource.TestCheckResourceAttr(name, prop("redis_maxmemory_policy"), "allkeys-lru"),
					resource.TestCheckResourceAttr(name, prop("redis_persistence"), "off"),
					resource.TestCheckResourceAttr(name, prop("redis_timeout"), "310"),
					resource.TestCheckResourceAttr(name, prop("redis_acl_channels_default"), "allchannels"),
					resource.TestCheckResourceAttr(name, prop("automatic_utility_network_ip_filter"), "false"),
				),
			},
			{
				Config: testDataS2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, prop("public_access"), "true"),
					resource.TestCheckResourceAttr(name, prop("redis_lfu_decay_time"), "1"),
					resource.TestCheckResourceAttr(name, prop("redis_number_of_databases"), "3"),
					resource.TestCheckResourceAttr(name, prop("redis_notify_keyspace_events"), ""),
					resource.TestCheckResourceAttr(name, prop("redis_pubsub_client_output_buffer_limit"), "256"),
					resource.TestCheckResourceAttr(name, prop("redis_ssl"), "true"),
					resource.TestCheckResourceAttr(name, prop("redis_lfu_log_factor"), "12"),
					resource.TestCheckResourceAttr(name, prop("redis_io_threads"), "1"),
					resource.TestCheckResourceAttr(name, prop("redis_maxmemory_policy"), "volatile-lru"),
					resource.TestCheckResourceAttr(name, prop("redis_persistence"), "rdb"),
					resource.TestCheckResourceAttr(name, prop("redis_timeout"), "320"),
					resource.TestCheckResourceAttr(name, prop("redis_acl_channels_default"), "resetchannels"),
					resource.TestCheckResourceAttr(name, prop("automatic_utility_network_ip_filter"), "true"),
				),
			},
		},
	})
}
