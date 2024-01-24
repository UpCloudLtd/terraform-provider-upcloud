resource "upcloud_managed_database_postgresql" "postgresql_properties" {
  name  = "postgresql-properties-test"
  title = "postgresql-properties-test"
  plan  = "1x1xCPU-2GB-25GB"
  zone  = "fi-hel1"
  properties {
    pg_stat_monitor_pgsm_max_buckets       = 10
    pg_stat_monitor_pgsm_enable_query_plan = true
    log_temp_files                         = 16
    pg_stat_monitor_enable                 = true
  }
}
