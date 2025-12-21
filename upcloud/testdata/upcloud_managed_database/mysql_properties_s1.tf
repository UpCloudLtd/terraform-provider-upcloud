variable "prefix" {
  default = "tf-acc-test-mysql-props-"
  type    = string
}

variable "zone" {
  default = "fi-hel2"
  type    = string
}

resource "upcloud_managed_database_mysql" "mysql_properties" {
  name  = "mysql-props-test"
  title = "${var.prefix}db"
  plan  = "1x1xCPU-2GB-25GB"
  zone  = var.zone
  properties {
    admin_username                      = "demoadmin"
    admin_password                      = "2VCNXEV6SVfpr3X1"
    automatic_utility_network_ip_filter = true
    backup_hour                         = 1
    backup_minute                       = 1
    binlog_retention_period             = 600
    connect_timeout                     = 2
    default_time_zone                   = "+02:00"
    group_concat_max_len                = 4
    information_schema_stats_expiry     = 900
    innodb_ft_min_token_size            = 1
    innodb_ft_server_stopword_table     = "db_name/table_name"
    innodb_lock_wait_timeout            = 1
    innodb_log_buffer_size              = 1048576
    innodb_online_alter_log_max_size    = 65536
    innodb_print_all_deadlocks          = true
    innodb_rollback_on_timeout          = true
    interactive_timeout                 = 30
    internal_tmp_mem_storage_engine     = "MEMORY"
    ip_filter                           = ["127.0.0.1/32", "127.0.0.2/32"]
    long_query_time                     = 1
    max_allowed_packet                  = 102400
    max_heap_table_size                 = 1048576
    net_read_timeout                    = 1
    net_write_timeout                   = 1
    public_access                       = false
    slow_query_log                      = true
    sort_buffer_size                    = 32768
    sql_mode                            = "ANSI,TRADITIONAL"
    sql_require_primary_key             = true
    tmp_table_size                      = 1048576
    version                             = "8"
    wait_timeout                        = 1
    service_log                         = true
  }
}
