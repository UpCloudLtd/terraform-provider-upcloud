# Create a network for the Kubernetes cluster
resource "upcloud_network" "example" {
  name = "example-network"
  zone = "de-fra1"
  ip_network {
    address = "172.16.1.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

# Create a Kubernetes cluster
resource "upcloud_kubernetes_cluster" "example" {
  name    = "exampleapp"
  network = upcloud_network.example.id
  zone    = "de-fra1"
}
