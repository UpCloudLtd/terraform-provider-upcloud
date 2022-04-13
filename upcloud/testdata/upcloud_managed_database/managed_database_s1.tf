resource "upcloud_managed_database_postgresql" "pg1" {
  name                    = "pg1"
  plan                    = "1x1xCPU-2GB-25GB"
  title                   = "tf-test-pg-1"
  zone                    = "pl-waw1"
  maintenance_window_time = "10:00:00"
  maintenance_window_dow  = "friday"
  powered                 = true
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
  powered = false
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

resource "upcloud_managed_database_logical_database" "logical_db_1" {
  service = upcloud_managed_database_mysql.msql1.id
  name    = "tf-test-logical-db-1"
}

resource "upcloud_managed_database_user" "db_user_1" {
  service  = upcloud_managed_database_mysql.msql1.id
  username = "somename"
  password = "Superpass123"
}
