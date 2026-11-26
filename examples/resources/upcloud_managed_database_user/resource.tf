resource "upcloud_managed_database_postgresql" "example" {
  name             = "postgres"
  plan_compute     = "rdb.standard.2CPU-8GB"
  plan_node_count  = 2
  plan_storage_gib = 120
  plan_backups     = "regular"
  title            = "postgres"
  zone             = "fi-hel1"
}

resource "upcloud_managed_database_user" "example_user" {
  service  = upcloud_managed_database_postgresql.example.id
  username = "example_user"
  password = "<USER_PASSWORD>"
}
