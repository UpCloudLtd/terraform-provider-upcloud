resource "upcloud_managed_database_postgresql" "example" {
  maintenance_window_time = "10:00:00"
  maintenance_window_dow  = "friday"
  name                    = "postgres"
  plan                    = "1x1xCPU-2GB-25GB"
  powered                 = true
  title                   = "postgres"
  zone                    = "pl-waw1"
}

resource "upcloud_managed_database_connection_pool" "example" {
  service  = upcloud_managed_database_postgresql.example.id
  database = "defaultdb"
  mode     = "session"
  name     = "example-2"
  size     = 95
}
