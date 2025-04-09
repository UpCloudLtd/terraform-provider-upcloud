variable "prefix" {
  default = "tf-acc-test-postgresql-props-"
  type    = string
}

variable "zone" {
  default = "fi-hel1"
  type    = string
}

resource "upcloud_managed_database_postgresql" "postgresql_properties" {
  name  = "postgresql-props-test"
  title = "${var.prefix}db"
  plan  = "1x1xCPU-2GB-25GB"
  zone  = var.zone
  properties {
    pg_stat_monitor_pgsm_max_buckets       = 10
    pg_stat_monitor_pgsm_enable_query_plan = true
    log_temp_files                         = 16
    pg_stat_monitor_enable                 = true
  }
}
