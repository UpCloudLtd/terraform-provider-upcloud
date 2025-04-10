variable "prefix" {
  default = "tf-acc-test-valkey-props-"
  type    = string
}

variable "zone" {
  default = "fi-hel2"
  type    = string
}

resource "upcloud_managed_database_valkey" "valkey_properties" {
  name  = "valkey-props-test"
  title = "${var.prefix}db"
  plan  = "1x1xCPU-2GB"
  zone  = var.zone
  properties {
    automatic_utility_network_ip_filter      = true
    public_access                            = true
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
