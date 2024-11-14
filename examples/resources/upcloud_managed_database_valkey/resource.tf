# Minimal config
resource "upcloud_managed_database_valkey" "example_1" {
  name  = "valkey-1"
  title = "valkey-example-1"
  plan  = "1x1xCPU-2GB"
  zone  = "fi-hel2"
}

# Service with custom properties
resource "upcloud_managed_database_valkey" "example_2" {
  name  = "valkey-2"
  title = "valkey-example-2"
  plan  = "1x1xCPU-2GB"
  zone  = "fi-hel1"

  properties {
    public_access = false
  }
}
