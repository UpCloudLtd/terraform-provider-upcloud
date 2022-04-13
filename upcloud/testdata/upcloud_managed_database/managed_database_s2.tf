resource "upcloud_managed_database_postgresql" "pg1" {
  name                    = "pg1"
  plan                    = "1x1xCPU-2GB-25GB"
  title                   = "tf-test-updated-pg-1"
  zone                    = "pl-waw1"
  maintenance_window_time = "11:00:00"
  maintenance_window_dow  = "thursday"
  powered                 = true
  properties {
    ip_filter = []
    version   = 14
  }
}

resource "upcloud_managed_database_postgresql" "pg2" {
  name    = "pg2"
  plan    = "1x1xCPU-2GB-25GB"
  title   = "tf-test-updated-pg-2"
  zone    = "pl-waw1"
  powered = false
  properties {
    version = 14
  }
}

resource "upcloud_managed_database_mysql" "msql1" {
  name    = "msql1"
  plan    = "1x1xCPU-2GB-25GB"
  title   = "tf-test-updated-msql-1"
  zone    = "pl-waw1"
  powered = false
}