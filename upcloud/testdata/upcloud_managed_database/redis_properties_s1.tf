resource "upcloud_managed_database_redis" "redis_properties" {
  name = "tf-redis-properties-test"
  plan = "2x4xCPU-28GB"
  zone = "fi-hel2"
  properties {
    automatic_utility_network_ip_filter     = false
    public_access                           = false
    redis_lfu_decay_time                    = 2
    redis_number_of_databases               = 2
    redis_notify_keyspace_events            = "KEA"
    redis_pubsub_client_output_buffer_limit = 128
    redis_ssl                               = false
    redis_lfu_log_factor                    = 11
    redis_io_threads                        = 2
    redis_maxmemory_policy                  = "allkeys-lru"
    redis_persistence                       = "off"
    redis_timeout                           = 310
    redis_acl_channels_default              = "allchannels"
  }
}
