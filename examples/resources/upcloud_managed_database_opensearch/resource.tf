# Minimal config
resource "upcloud_managed_database_opensearch" "example_1" {
  name  = "opensearch-1"
  title = "opensearch-1-example-1"
  plan  = "1x2xCPU-4GB-80GB-1D"
  zone  = "fi-hel2"
}

# Service with custom properties and access control
resource "upcloud_managed_database_opensearch" "example_2" {
  name                    = "opensearch-2"
  title                   = "opensearch-2-example-2"
  plan                    = "1x2xCPU-4GB-80GB-1D"
  zone                    = "fi-hel1"
  access_control          = true
  extended_access_control = true
  properties {
    public_access = false
  }
}
