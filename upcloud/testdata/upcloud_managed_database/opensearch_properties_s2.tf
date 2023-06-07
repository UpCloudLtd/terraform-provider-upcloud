resource "upcloud_managed_database_opensearch" "opensearch_properties" {
  name                    = "opensearch-properties-test-2"
  plan                    = "1x2xCPU-4GB-80GB-1D"
  zone                    = "fi-hel2"
  access_control          = true
  extended_access_control = true
  properties {
    automatic_utility_network_ip_filter = true
    public_access                       = true
  }
}
