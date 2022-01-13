resource "upcloud_managed_database_postgresql" "example" {
  name  = "postgres"
  plan  = "1x1xCPU-2GB-25GB"
  title = "postgres"
  zone  = "fi-hel1"
}

resource "upcloud_managed_database_user" "example_user" {
  service  = upcloud_managed_database_postgresql.example.id
  username = "example_user"
  password = "<USER_PASSWORD>"
}
