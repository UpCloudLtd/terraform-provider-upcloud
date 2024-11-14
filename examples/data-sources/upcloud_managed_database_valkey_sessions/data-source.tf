# Use data source to gather a list of the active sessions for a Managed Valkey Database

# Create a Managed Valkey resource
resource "upcloud_managed_database_valkey" "example" {
  name  = "example"
  title = "example"
  plan  = "1x1xCPU-2GB"
  zone  = "fi-hel2"
}

# Read the active sessions of the newly created service
data "upcloud_managed_database_valkey_sessions" "example" {
  service = upcloud_managed_database_valkey.example.id
}
