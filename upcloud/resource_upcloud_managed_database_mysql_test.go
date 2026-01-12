package upcloud

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUpcloudManagedDatabaseMySQLProperties(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/mysql_properties_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/mysql_properties_s2.tf")

	name := "upcloud_managed_database_mysql.mysql_properties"
	prop := func(name string) string {
		return fmt.Sprintf("properties.0.%s", name)
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(name, "zone", "fi-hel2"),
					resource.TestCheckResourceAttr(name, prop("default_time_zone"), "+02:00"),
					resource.TestCheckResourceAttr(name, prop("admin_username"), "demoadmin"),
					resource.TestCheckResourceAttr(name, prop("admin_password"), "2VCNXEV6SVfpr3X1"),
					resource.TestCheckResourceAttr(name, prop("automatic_utility_network_ip_filter"), "true"),
					resource.TestCheckResourceAttr(name, prop("backup_hour"), "1"),
					resource.TestCheckResourceAttr(name, prop("backup_minute"), "1"),
					resource.TestCheckResourceAttr(name, prop("ip_filter.0"), "127.0.0.1/32"),
					resource.TestCheckResourceAttr(name, prop("ip_filter.1"), "127.0.0.2/32"),
					resource.TestCheckResourceAttr(name, prop("binlog_retention_period"), "600"),
					resource.TestCheckResourceAttr(name, prop("connect_timeout"), "2"),
					resource.TestCheckResourceAttr(name, prop("group_concat_max_len"), "4"),
					resource.TestCheckResourceAttr(name, prop("information_schema_stats_expiry"), "900"),
					resource.TestCheckResourceAttr(name, prop("innodb_ft_min_token_size"), "1"),
					resource.TestCheckResourceAttr(name, prop("innodb_ft_server_stopword_table"), "db_name/table_name"),
					resource.TestCheckResourceAttr(name, prop("innodb_lock_wait_timeout"), "1"),
					resource.TestCheckResourceAttr(name, prop("innodb_log_buffer_size"), "1048576"),
					resource.TestCheckResourceAttr(name, prop("innodb_online_alter_log_max_size"), "65536"),
					resource.TestCheckResourceAttr(name, prop("innodb_print_all_deadlocks"), "true"),
					resource.TestCheckResourceAttr(name, prop("innodb_rollback_on_timeout"), "true"),
					resource.TestCheckResourceAttr(name, prop("interactive_timeout"), "30"),
					resource.TestCheckResourceAttr(name, prop("internal_tmp_mem_storage_engine"), "MEMORY"),
					resource.TestCheckResourceAttr(name, prop("long_query_time"), "1"),
					resource.TestCheckResourceAttr(name, prop("max_allowed_packet"), "102400"),
					resource.TestCheckResourceAttr(name, prop("max_heap_table_size"), "1048576"),
					resource.TestCheckResourceAttr(name, prop("net_read_timeout"), "1"),
					resource.TestCheckResourceAttr(name, prop("net_write_timeout"), "1"),
					resource.TestCheckResourceAttr(name, prop("public_access"), "false"),
					resource.TestCheckResourceAttr(name, prop("slow_query_log"), "true"),
					resource.TestCheckResourceAttr(name, prop("sort_buffer_size"), "32768"),
					resource.TestCheckResourceAttr(name, prop("sql_mode"), "ANSI,TRADITIONAL"),
					resource.TestCheckResourceAttr(name, prop("sql_require_primary_key"), "true"),
					resource.TestCheckResourceAttr(name, prop("tmp_table_size"), "1048576"),
					resource.TestCheckResourceAttr(name, prop("version"), "8"),
					resource.TestCheckResourceAttr(name, prop("wait_timeout"), "1"),
					resource.TestCheckResourceAttr(name, prop("service_log"), "true"),
					// there should be mysqlx and mysql component
					resource.TestCheckResourceAttr(name, "components.#", "2"),
					resource.TestCheckResourceAttr(name, "node_states.0.state", "running"),
					resource.TestCheckResourceAttr(name, "node_states.0.role", "master"),
				),
			},
			{
				Config: testDataS2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, prop("admin_username"), "demoadmin"),
					resource.TestCheckResourceAttr(name, prop("admin_password"), "2VCNXEV6SVfpr3X1"),
					resource.TestCheckResourceAttr(name, prop("innodb_read_io_threads"), "10"),
					resource.TestCheckResourceAttr(name, prop("innodb_flush_neighbors"), "0"),
					resource.TestCheckResourceAttr(name, prop("innodb_change_buffer_max_size"), "26"),
					resource.TestCheckResourceAttr(name, prop("net_buffer_length"), "1024"),
					resource.TestCheckResourceAttr(name, prop("innodb_thread_concurrency"), "2"),
					resource.TestCheckResourceAttr(name, prop("innodb_write_io_threads"), "5"),
				),
			},
		},
	})
}
