# Use data source to gather a list of the indices for a Managed OpenSearch Database

# Create a Managed OpenSearch resource
resource "upcloud_managed_database_opensearch" "example" {
  name = "opensearch-example"
  plan = "1x2xCPU-4GB-80GB-1D"
  zone = "fi-hel1"
  properties {
    automatic_utility_network_ip_filter = false
    public_access                       = false
  }
}

# Read the available indices of the newly created service
data "upcloud_managed_database_opensearch_indices" "example" {
  service = upcloud_managed_database_opensearch.example.id
}
