# Use Kubernetes provider to access your Kubernetes cluster

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
  control_plane_ip_filter = ["0.0.0.0/0"]
  name                    = "exampleapp"
  network                 = upcloud_network.example.id
  zone                    = "de-fra1"
}

# Read the details of the newly created cluster
data "upcloud_kubernetes_cluster" "example" {
  id = upcloud_kubernetes_cluster.example.id
}

# Set the Kubernetes provider credentials
provider "kubernetes" {
  client_certificate     = data.upcloud_kubernetes_cluster.example.client_certificate
  client_key             = data.upcloud_kubernetes_cluster.example.client_key
  cluster_ca_certificate = data.upcloud_kubernetes_cluster.example.cluster_ca_certificate
  host                   = data.upcloud_kubernetes_cluster.example.host
}

# Use the Kubernetes provider resources to interact with the cluster
resource "kubernetes_namespace" "example" {
  metadata {
    name = "example-namespace"
  }
}

# In addition, write the kubeconfig to a file to interact with the cluster with `kubectl` or other clients
resource "local_file" "example" {
  content  = data.upcloud_kubernetes_cluster.example.kubeconfig
  filename = "example.conf"
}
