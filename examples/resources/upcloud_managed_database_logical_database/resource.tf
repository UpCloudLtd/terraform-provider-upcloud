# PostgreSQL managed database with additional logical database: example_db 
resource "upcloud_managed_database_postgresql" "example" {
  name             = "postgres"
  plan_compute     = "rdb.standard.2CPU-8GB"
  plan_node_count  = 2
  plan_storage_gib = 120
  plan_backups     = "regular"
  title            = "postgres"
  zone             = "fi-hel1"
}

resource "upcloud_managed_database_logical_database" "example_db" {
  service = upcloud_managed_database_postgresql.example.id
  name    = "example_db"
}

# MySQL managed database with additional logical database: example2_db 
resource "upcloud_managed_database_mysql" "example" {
  name             = "mysql"
  plan_compute     = "rdb.standard.2CPU-8GB"
  plan_node_count  = 2
  plan_storage_gib = 120
  plan_backups     = "regular"
  title            = "mysql"
  zone             = "fi-hel1"
}

resource "upcloud_managed_database_logical_database" "example2_db" {
  service = upcloud_managed_database_mysql.example.id
  name    = "example2_db"
}
