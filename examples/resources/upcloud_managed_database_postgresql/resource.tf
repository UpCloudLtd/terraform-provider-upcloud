# Minimal config
resource "upcloud_managed_database_postgresql" "example_1" {
  name             = "postgres-1"
  plan_compute     = "rdb.standard.2CPU-8GB"
  plan_node_count  = 2
  plan_storage_gib = 120
  plan_backups     = "regular"
  title            = "postgres"
  zone             = "fi-hel1"
}

# Service with custom properties
resource "upcloud_managed_database_postgresql" "example_2" {
  name             = "postgres-2"
  plan_compute     = "rdb.standard.2CPU-8GB"
  plan_node_count  = 2
  plan_storage_gib = 120
  plan_backups     = "regular"
  title            = "postgres"
  zone             = "fi-hel1"
  properties {
    timezone       = "Europe/Helsinki"
    admin_username = "admin"
    admin_password = "<ADMIN_PASSWORD>"
  }
}
