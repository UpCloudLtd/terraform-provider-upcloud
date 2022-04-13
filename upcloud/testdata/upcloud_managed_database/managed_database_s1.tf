resource "upcloud_managed_database_postgresql" "pg1" {
  name                    = "pg1"
  plan                    = "1x1xCPU-2GB-25GB"
  title                   = "tf-test-pg-1"
  zone                    = "pl-waw1"
  maintenance_window_time = "10:00:00"
  maintenance_window_dow  = "friday"
  powered                 = false
  properties {
    public_access = true
    ip_filter     = ["10.0.0.1/32"]
    version       = 13
  }
}

resource "upcloud_managed_database_postgresql" "pg2" {
  name    = "pg2"
  plan    = "1x1xCPU-2GB-25GB"
  title   = "tf-test-pg-2"
  zone    = "pl-waw1"
  powered = true
  properties {
    version = 13
  }
}

resource "upcloud_managed_database_mysql" "msql1" {
  name  = "msql1"
  plan  = "1x1xCPU-2GB-25GB"
  title = "tf-test-msql-1"
  zone  = "pl-waw1"
}