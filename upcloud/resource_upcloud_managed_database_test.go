package upcloud

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudManagedDatabase(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/managed_database_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/managed_database_s2.tf")

	var providers []*schema.Provider
	pg1Name := "upcloud_managed_database_postgresql.pg1"
	pg2Name := "upcloud_managed_database_postgresql.pg2"
	msql1Name := "upcloud_managed_database_mysql.msql1"
	lgDBName := "upcloud_managed_database_logical_database.logical_db_1"
	userName1 := "upcloud_managed_database_user.db_user_1"
	userName2 := "upcloud_managed_database_user.db_user_2"
	userName3 := "upcloud_managed_database_user.db_user_3"
	userName4 := "upcloud_managed_database_user.db_user_4"
	redisName := "upcloud_managed_database_redis.r1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(pg1Name, "name", "tf-pg-test-1"),
					resource.TestCheckResourceAttr(pg1Name, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(pg1Name, "title", "tf-test-pg-1"),
					resource.TestCheckResourceAttr(pg1Name, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(pg1Name, "powered", "true"),
					resource.TestCheckResourceAttr(pg1Name, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(pg1Name, "maintenance_window_dow", "friday"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.ip_filter.0", "10.0.0.1/32"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.version", "13"),
					resource.TestCheckResourceAttr(pg1Name, "type", string(upcloud.ManagedDatabaseServiceTypePostgreSQL)),
					resource.TestCheckResourceAttrSet(pg1Name, "service_uri"),

					resource.TestCheckResourceAttr(pg2Name, "name", "tf-pg-test-2"),
					resource.TestCheckResourceAttr(pg2Name, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(pg2Name, "title", "tf-test-pg-2"),
					resource.TestCheckResourceAttr(pg2Name, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(pg2Name, "powered", "true"),
					resource.TestCheckResourceAttr(pg2Name, "properties.0.version", "14"),

					resource.TestCheckResourceAttr(msql1Name, "name", "tf-mysql-test-2"),
					resource.TestCheckResourceAttr(msql1Name, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(msql1Name, "title", "tf-test-msql-1"),
					resource.TestCheckResourceAttr(msql1Name, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(msql1Name, "powered", "true"),

					resource.TestCheckResourceAttr(lgDBName, "name", "tf-test-logical-db-1"),
					resource.TestCheckResourceAttrSet(lgDBName, "service"),

					resource.TestCheckResourceAttr(userName1, "username", "somename"),
					resource.TestCheckResourceAttr(userName1, "password", "Superpass123"),
					resource.TestCheckResourceAttr(userName1, "authentication", "mysql_native_password"),
					resource.TestCheckResourceAttrSet(userName1, "service"),

					resource.TestCheckResourceAttr(userName2, "pg_access_control.0.allow_replication", "false"),

					resource.TestCheckResourceAttr(userName3, "redis_access_control.0.categories.0", "+@set"),
					resource.TestCheckResourceAttr(userName3, "redis_access_control.0.channels.0", "*"),
					resource.TestCheckResourceAttr(userName3, "redis_access_control.0.commands.0", "+set"),
					resource.TestCheckResourceAttr(userName3, "redis_access_control.0.keys.0", "key_*"),

					resource.TestCheckResourceAttr(userName4, "opensearch_access_control.0.rules.0.index", ".opensearch-observability"),
					resource.TestCheckResourceAttr(userName4, "opensearch_access_control.0.rules.0.permission", "admin"),

					resource.TestCheckResourceAttr(redisName, "name", "tf-redis-test-1"),
					resource.TestCheckResourceAttr(redisName, "plan", "1x1xCPU-2GB"),
					resource.TestCheckResourceAttr(redisName, "title", "tf-test-redis-1"),
					resource.TestCheckResourceAttr(redisName, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(redisName, "powered", "true"),
				),
			},
			{
				Config: testDataS2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(pg1Name, "title", "tf-test-updated-pg-1"),
					resource.TestCheckResourceAttr(pg1Name, "maintenance_window_time", "11:00:00"),
					resource.TestCheckResourceAttr(pg1Name, "maintenance_window_dow", "thursday"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.public_access", "false"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.ip_filter.#", "0"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.version", "14"),
					resource.TestCheckResourceAttr(pg1Name, "powered", "false"),

					resource.TestCheckResourceAttr(pg2Name, "title", "tf-test-updated-pg-2"),
					resource.TestCheckResourceAttr(pg2Name, "powered", "true"),
					resource.TestCheckResourceAttr(pg2Name, "properties.0.version", "14"),

					resource.TestCheckResourceAttr(msql1Name, "title", "tf-test-updated-msql-1"),

					resource.TestCheckResourceAttr(lgDBName, "name", "tf-test-updated-logical-db-1"),

					resource.TestCheckResourceAttr(userName1, "password", "Superpass890"),
					resource.TestCheckResourceAttr(userName1, "authentication", "caching_sha2_password"),

					resource.TestCheckResourceAttr(userName2, "pg_access_control.0.allow_replication", "true"),

					resource.TestCheckResourceAttr(userName3, "redis_access_control.0.categories.#", "0"),
					resource.TestCheckResourceAttr(userName3, "redis_access_control.0.channels.#", "0"),
					resource.TestCheckResourceAttr(userName3, "redis_access_control.0.commands.#", "0"),
					resource.TestCheckResourceAttr(userName3, "redis_access_control.0.keys.0", "key*"),

					resource.TestCheckResourceAttr(redisName, "name", "tf-redis-test-1"),
					resource.TestCheckResourceAttr(redisName, "plan", "1x1xCPU-2GB"),
					resource.TestCheckResourceAttr(redisName, "title", "tf-test-redis-title-1"),
					resource.TestCheckResourceAttr(redisName, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(redisName, "powered", "true"),
				),
			},
		},
	})
}
