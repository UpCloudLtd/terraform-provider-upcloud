resource "upcloud_managed_database_valkey" "valkey_properties" {
  name  = "tf-valkey-properties-test"
  title = "tf-valkey-properties-test"
  plan  = "1x1xCPU-2GB"
  zone  = "fi-hel2"
  properties {
    automatic_utility_network_ip_filter     = false
    public_access                           = false
    valkey_lfu_decay_time                    = 2
    valkey_number_of_databases               = 2
    valkey_notify_keyspace_events            = "KEA"
    valkey_pubsub_client_output_buffer_limit = 128
    valkey_ssl                               = false
    valkey_lfu_log_factor                    = 11
    valkey_io_threads                        = 1
    valkey_maxmemory_policy                  = "allkeys-lru"
    valkey_persistence                       = "off"
    valkey_timeout                           = 310
    valkey_acl_channels_default              = "allchannels"
    service_log                             = true
  }
}
