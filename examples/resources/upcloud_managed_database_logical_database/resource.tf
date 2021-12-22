resource "upcloud_managed_database_postgresql" "example" {
  name  = "postgres"
  plan  = "1x1xCPU-2GB-25GB"
  title = "postgres"
  zone  = "fi-hel1"
}

resource "upcloud_managed_database_logical_database" "example_db" {
  service = upcloud_managed_database_postgresql.example.id
  name    = "example_db"
}
