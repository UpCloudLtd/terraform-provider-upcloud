resource "upcloud_managed_database_mysql" "mysql_properties" {
  name = "mysql-properties-test"
  plan = "1x1xCPU-2GB-25GB"
  zone = "fi-hel2"
  properties {
    innodb_read_io_threads        = 10
    innodb_flush_neighbors        = 0
    innodb_change_buffer_max_size = 26
    net_buffer_length             = 1024
    innodb_thread_concurrency     = 2
    innodb_write_io_threads       = 5
  }
}
