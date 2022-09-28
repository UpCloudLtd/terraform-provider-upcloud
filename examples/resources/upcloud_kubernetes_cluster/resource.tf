# Use `upcloud_kubernetes_plan` to query node group plans
data "upcloud_kubernetes_plan" "medium" {
  name = "medium"
}

data "upcloud_kubernetes_plan" "large" {
  name = "large"
}

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
      plan = data.upcloud_kubernetes_plan.large.description
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
      plan = data.upcloud_kubernetes_plan.medium.description
      ssh_keys = [
        // insert your SSH key(s) here to access the worker nodes
      ]
    }
  ]
  zone = upcloud_network.example.zone
}
