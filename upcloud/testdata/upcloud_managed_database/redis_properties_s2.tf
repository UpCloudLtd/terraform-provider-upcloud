resource "upcloud_managed_database_redis" "redis_properties" {
  name = "tf-redis-properties-test"
  plan = "1x1xCPU-2GB"
  zone = "fi-hel2"
  properties {
    automatic_utility_network_ip_filter     = true
    public_access                           = true
    redis_lfu_decay_time                    = 1
    redis_number_of_databases               = 3
    redis_notify_keyspace_events            = ""
    redis_pubsub_client_output_buffer_limit = 256
    redis_ssl                               = true
    redis_lfu_log_factor                    = 12
    redis_io_threads                        = 1
    redis_maxmemory_policy                  = "volatile-lru"
    redis_persistence                       = "rdb"
    redis_timeout                           = 320
    redis_acl_channels_default              = "resetchannels"
  }
}
