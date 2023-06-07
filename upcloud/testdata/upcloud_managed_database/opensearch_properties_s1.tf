resource "upcloud_managed_database_opensearch" "opensearch_properties" {
  name = "opensearch-properties-test-1"
  plan = "1x2xCPU-4GB-80GB-1D"
  zone = "fi-hel2"
  properties {
    automatic_utility_network_ip_filter = false
    public_access                       = false
    version                             = "1"
  }
}

data "upcloud_managed_database_opensearch_indices" "opensearch_indices" {
  service = upcloud_managed_database_opensearch.opensearch_properties.id
}
