# Use data source to gather a list of the active sessions for a Managed PostgreSQL Database

# Create a Managed PostgreSQL resource
resource "upcloud_managed_database_postgresql" "example" {
  name             = "postgresql-example1"
  title            = "postgresql-example1"
  plan_compute     = "rdb.standard.2CPU-8GB"
  plan_node_count  = 2
  plan_storage_gib = 120
  plan_backups     = "regular"
  zone             = "fi-hel1"
}

# Read the active sessions of the newly created service
data "upcloud_managed_database_postgresql_sessions" "example" {
  service = upcloud_managed_database_postgresql.example.id
}
