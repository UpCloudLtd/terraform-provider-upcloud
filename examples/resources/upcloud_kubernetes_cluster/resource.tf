# Use `upcloud_kubernetes_plans` to query node group plans
data "upcloud_kubernetes_plans" "example" {}

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
  name    = "example"
  network = upcloud_network.example.id
  node_groups = [
    {
      count = 4
      labels = {
        managedBy = "terraform"
      }
      name = "node-group-large"
      plan = lookup(data.upcloud_kubernetes_plans.example.plans, "large", "K8S-4xCPU-8GB")
      ssh_keys = [
        // insert your SSH key(s) here to access the worker nodes
      ]
    },
    {
      count = 8
      labels = {
        managedBy = "terraform"
      }
      name = "node-group-medium"
      plan = lookup(data.upcloud_kubernetes_plans.example.plans, "medium", "K8S-4xCPU-8GB")
      ssh_keys = [
        // insert your SSH key(s) here to access the worker nodes
      ]
    }
  ]
  zone = upcloud_network.example.zone
}
