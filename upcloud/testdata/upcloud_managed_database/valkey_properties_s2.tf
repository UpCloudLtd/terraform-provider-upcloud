resource "upcloud_managed_database_valkey" "valkey_properties" {
  name  = "tf-valkey-properties-test"
  title = "tf-valkey-properties-test"
  plan  = "1x1xCPU-2GB"
  zone  = "fi-hel2"
  properties {
    automatic_utility_network_ip_filter     = true
    public_access                           = true
    valkey_lfu_decay_time                    = 1
    valkey_number_of_databases               = 3
    valkey_notify_keyspace_events            = ""
    valkey_pubsub_client_output_buffer_limit = 256
    valkey_ssl                               = true
    valkey_lfu_log_factor                    = 12
    valkey_io_threads                        = 1
    valkey_maxmemory_policy                  = "volatile-lru"
    valkey_persistence                       = "rdb"
    valkey_timeout                           = 320
    valkey_acl_channels_default              = "resetchannels"
  }
}
