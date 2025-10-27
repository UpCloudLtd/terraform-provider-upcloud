package upcloud

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUpcloudManagedDatabasePostgreSQLProperties(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/postgresql_properties_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/postgresql_properties_s2.tf")

	name := "upcloud_managed_database_postgresql.postgresql_properties"
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
					resource.TestCheckResourceAttr(name, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(name, "zone", "fi-hel1"),
					resource.TestCheckResourceAttr(name, prop("timezone"), "Europe/Helsinki"),
					resource.TestCheckResourceAttr(name, prop("admin_username"), "demoadmin"),
					resource.TestCheckResourceAttr(name, prop("admin_password"), "2VCNXEV6SVfpr3"),
					resource.TestCheckResourceAttr(name, prop("automatic_utility_network_ip_filter"), "true"),
					resource.TestCheckResourceAttr(name, prop("autovacuum_analyze_scale_factor"), "0.1"),
					resource.TestCheckResourceAttr(name, prop("autovacuum_analyze_threshold"), "1"),
					resource.TestCheckResourceAttr(name, prop("autovacuum_freeze_max_age"), "200000000"),
					resource.TestCheckResourceAttr(name, prop("autovacuum_max_workers"), "1"),
					resource.TestCheckResourceAttr(name, prop("autovacuum_naptime"), "1"),
					resource.TestCheckResourceAttr(name, prop("autovacuum_vacuum_cost_delay"), "1"),
					resource.TestCheckResourceAttr(name, prop("autovacuum_vacuum_cost_limit"), "1"),
					resource.TestCheckResourceAttr(name, prop("autovacuum_vacuum_scale_factor"), "0.2"),
					resource.TestCheckResourceAttr(name, prop("autovacuum_vacuum_threshold"), "1"),
					resource.TestCheckResourceAttr(name, prop("backup_hour"), "1"),
					resource.TestCheckResourceAttr(name, prop("backup_minute"), "1"),
					resource.TestCheckResourceAttr(name, prop("bgwriter_delay"), "10"),
					resource.TestCheckResourceAttr(name, prop("bgwriter_flush_after"), "1"),
					resource.TestCheckResourceAttr(name, prop("bgwriter_lru_maxpages"), "1"),
					resource.TestCheckResourceAttr(name, prop("bgwriter_lru_multiplier"), "9.2"),
					resource.TestCheckResourceAttr(name, prop("deadlock_timeout"), "501"),
					resource.TestCheckResourceAttr(name, prop("default_toast_compression"), "lz4"),
					resource.TestCheckResourceAttr(name, prop("idle_in_transaction_session_timeout"), "1000"),
					resource.TestCheckResourceAttr(name, prop("ip_filter.0"), "127.0.0.1/32"),
					resource.TestCheckResourceAttr(name, prop("ip_filter.1"), "127.0.0.2/32"),
					resource.TestCheckResourceAttr(name, prop("jit"), "true"),
					resource.TestCheckResourceAttr(name, prop("log_autovacuum_min_duration"), "1"),
					resource.TestCheckResourceAttr(name, prop("log_error_verbosity"), "DEFAULT"),
					resource.TestCheckResourceAttr(name, prop("log_line_prefix"), "'%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '"),
					resource.TestCheckResourceAttr(name, prop("log_min_duration_statement"), "1"),
					resource.TestCheckResourceAttr(name, prop("max_files_per_process"), "1000"),
					resource.TestCheckResourceAttr(name, prop("max_locks_per_transaction"), "64"),
					resource.TestCheckResourceAttr(name, prop("max_logical_replication_workers"), "4"),
					resource.TestCheckResourceAttr(name, prop("max_parallel_workers"), "1"),
					resource.TestCheckResourceAttr(name, prop("max_parallel_workers_per_gather"), "1"),
					resource.TestCheckResourceAttr(name, prop("max_pred_locks_per_transaction"), "64"),
					resource.TestCheckResourceAttr(name, prop("max_prepared_transactions"), "1"),
					resource.TestCheckResourceAttr(name, prop("max_replication_slots"), "8"),
					resource.TestCheckResourceAttr(name, prop("max_slot_wal_keep_size"), "10"),
					resource.TestCheckResourceAttr(name, prop("max_stack_depth"), "2097152"),
					resource.TestCheckResourceAttr(name, prop("max_standby_archive_delay"), "1"),
					resource.TestCheckResourceAttr(name, prop("max_standby_streaming_delay"), "1"),
					resource.TestCheckResourceAttr(name, prop("max_wal_senders"), "20"),
					resource.TestCheckResourceAttr(name, prop("max_worker_processes"), "8"),
					resource.TestCheckResourceAttr(name, prop("public_access"), "false"),
					resource.TestCheckResourceAttr(name, prop("shared_buffers_percentage"), "20"),
					resource.TestCheckResourceAttr(name, prop("synchronous_replication"), ""),
					resource.TestCheckResourceAttr(name, prop("temp_file_limit"), "1"),
					resource.TestCheckResourceAttr(name, prop("track_activity_query_size"), "1024"),
					resource.TestCheckResourceAttr(name, prop("track_commit_timestamp"), "on"),
					resource.TestCheckResourceAttr(name, prop("track_functions"), "all"),
					resource.TestCheckResourceAttr(name, prop("track_io_timing"), "on"),
					resource.TestCheckResourceAttr(name, prop("variant"), ""),
					resource.TestCheckResourceAttr(name, prop("version"), "16"),
					resource.TestCheckResourceAttr(name, prop("pg_partman_bgw_interval"), "3600"),
					resource.TestCheckResourceAttr(name, prop("pg_partman_bgw_role"), "upadmin"),
					resource.TestCheckResourceAttr(name, prop("pg_stat_statements_track"), "all"),
					resource.TestCheckResourceAttr(name, prop("pgbouncer.0.autodb_idle_timeout"), "10"),
					resource.TestCheckResourceAttr(name, prop("pgbouncer.0.autodb_max_db_connections"), "5"),
					resource.TestCheckResourceAttr(name, prop("pgbouncer.0.autodb_pool_mode"), "session"),
					resource.TestCheckResourceAttr(name, prop("pgbouncer.0.autodb_pool_size"), "1"),
					resource.TestCheckResourceAttr(name, prop("pgbouncer.0.ignore_startup_parameters.0"), "search_path"),
					resource.TestCheckResourceAttr(name, prop("pgbouncer.0.min_pool_size"), "1"),
					resource.TestCheckResourceAttr(name, prop("pgbouncer.0.server_idle_timeout"), "10"),
					resource.TestCheckResourceAttr(name, prop("pgbouncer.0.server_lifetime"), "60"),
					resource.TestCheckResourceAttr(name, prop("pgbouncer.0.server_reset_query_always"), "false"),
					resource.TestCheckResourceAttr(name, prop("pglookout.0.max_failover_replication_time_lag"), "10"),
					resource.TestCheckResourceAttr(name, prop("timescaledb.0.max_background_workers"), "1"),
					resource.TestCheckResourceAttr(name, prop("service_log"), "true"),
					// there should be pgbouncer and pg component
					resource.TestCheckResourceAttr(name, "components.#", "2"),
					resource.TestCheckResourceAttr(name, "node_states.0.state", "running"),
					resource.TestCheckResourceAttr(name, "node_states.0.role", "master"),
				),
			},
			{
				Config:                  testDataS1,
				ResourceName:            name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"properties.0.admin_password", "properties.0.admin_username", "state"}, // credentials only provided on creation, not available on subsequent requests like import
			},
			{
				Config: testDataS2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, prop("admin_username"), "demoadmin"),
					resource.TestCheckResourceAttr(name, prop("admin_password"), "2VCNXEV6SVfpr3"),
					resource.TestCheckResourceAttr(name, prop("pg_stat_monitor_pgsm_max_buckets"), "10"),
					resource.TestCheckResourceAttr(name, prop("pg_stat_monitor_pgsm_enable_query_plan"), "true"),
					resource.TestCheckResourceAttr(name, prop("log_temp_files"), "16"),
					resource.TestCheckResourceAttr(name, prop("pg_stat_monitor_enable"), "true"),
					resource.TestCheckResourceAttr(name, prop("version"), "17"),
				),
			},
		},
	})
}
