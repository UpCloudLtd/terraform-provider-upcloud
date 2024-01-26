resource "upcloud_managed_database_postgresql" "postgresql_properties" {
  name  = "postgresql-properties-test"
  title = "postgresql-properties-test"
  plan  = "1x1xCPU-2GB-25GB"
  zone  = "fi-hel1"
  properties {
    timezone                            = "Europe/Helsinki"
    admin_username                      = "demoadmin"
    admin_password                      = "2VCNXEV6SVfpr3"
    automatic_utility_network_ip_filter = true
    autovacuum_analyze_scale_factor     = 0.1
    autovacuum_analyze_threshold        = 1
    autovacuum_freeze_max_age           = 200000000
    autovacuum_max_workers              = 1
    autovacuum_naptime                  = 1
    autovacuum_vacuum_cost_delay        = 1
    autovacuum_vacuum_cost_limit        = 1
    autovacuum_vacuum_scale_factor      = 0.2
    autovacuum_vacuum_threshold         = 1
    backup_hour                         = 1
    backup_minute                       = 1
    bgwriter_delay                      = 10
    bgwriter_flush_after                = 1
    bgwriter_lru_maxpages               = 1
    bgwriter_lru_multiplier             = 9.2
    deadlock_timeout                    = 501
    default_toast_compression           = "lz4"
    idle_in_transaction_session_timeout = 1
    ip_filter                           = ["127.0.0.1", "127.0.0.2"]
    jit                                 = true
    log_autovacuum_min_duration         = 1
    log_error_verbosity                 = "DEFAULT"
    log_line_prefix                     = "'%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '"
    log_min_duration_statement          = 1
    max_files_per_process               = 1000
    max_locks_per_transaction           = 64
    max_logical_replication_workers     = 4
    max_parallel_workers                = 1
    max_parallel_workers_per_gather     = 1
    max_pred_locks_per_transaction      = 64
    max_prepared_transactions           = 1
    max_replication_slots               = 8
    max_slot_wal_keep_size              = 10
    max_stack_depth                     = 2097152
    max_standby_archive_delay           = 1
    max_standby_streaming_delay         = 1
    max_wal_senders                     = 20
    max_worker_processes                = 8
    public_access                       = false
    shared_buffers_percentage           = 20
    temp_file_limit                     = 1
    track_activity_query_size           = 1024
    track_commit_timestamp              = "on"
    track_functions                     = "all"
    track_io_timing                     = "on"
    version                             = "15"
    wal_sender_timeout                  = 60000
    wal_writer_delay                    = 10
    work_mem                            = 1024
    pg_partman_bgw_interval             = 3600
    pg_partman_bgw_role                 = "upadmin"
    pg_stat_statements_track            = "all"
    pgbouncer {
      autodb_idle_timeout       = 1
      autodb_max_db_connections = 1
      autodb_pool_mode          = "session"
      autodb_pool_size          = 1
      ignore_startup_parameters = ["search_path"]
      min_pool_size             = 1
      server_idle_timeout       = 1
      server_lifetime           = 60
      server_reset_query_always = false
    }

    pglookout {
      max_failover_replication_time_lag = 10
    }

    timescaledb {
      max_background_workers = 1
    }

    service_log = true
  }
}
