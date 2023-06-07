resource "upcloud_managed_database_opensearch" "opensearch_indices" {
  name = "opensearch-indices-test-1"
  plan = "1x2xCPU-4GB-80GB-1D"
  zone = "fi-hel2"
  properties {
    automatic_utility_network_ip_filter = false
    public_access                       = false
  }
}

data "upcloud_managed_database_opensearch_indices" "opensearch_indices" {
  service = upcloud_managed_database_opensearch.opensearch_indices.id
}
