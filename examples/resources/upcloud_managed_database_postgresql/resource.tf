# Minimal config
resource "upcloud_managed_database_postgresql" "example_1" {
  name  = "postgres-1"
  plan  = "1x1xCPU-2GB-25GB"
  title = "postgres"
  zone  = "fi-hel1"
}

# Service with custom properties
resource "upcloud_managed_database_postgresql" "example_2" {
  name  = "postgres-2"
  plan  = "1x1xCPU-2GB-25GB"
  title = "postgres"
  zone  = "fi-hel1"
  properties {
    timezone       = "Europe/Helsinki"
    admin_username = "admin"
    admin_password = "<ADMIN_PASSWORD>"
  }
}
