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
  name    = "exampleapp"
  network = upcloud_network.example.id
  zone    = "de-fra1"

  # Create a group of worker nodes with common settings
  node_group {
    count    = 2
    name     = "group1"
    plan     = "K8S-2xCPU-4GB"
    ssh_keys = ["public_ssh_key"]

    labels = {
      managedBy = "terraform"
    }

    taint {
      effect = "NoExecute"
      key    = "taintKey"
      value  = "taintValue"
    }
  }

  # Create another node group, only required attributes
  node_group {
    count = 2
    name  = "group2"
    plan  = "K8S-2xCPU-4GB"
  }
}
