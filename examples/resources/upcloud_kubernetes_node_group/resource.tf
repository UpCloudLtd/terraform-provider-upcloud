# Create a network and a router for the Kubernetes cluster
resource "upcloud_router" "example" {
  name = "example-router"
}

resource "upcloud_network" "example" {
  name   = "example-network"
  zone   = "fi-hel2"
  router = upcloud_router.example.id

  ip_network {
    address = "172.16.1.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

# Create a Kubernetes cluster
resource "upcloud_kubernetes_cluster" "example" {
  # Allow access to the cluster control plane from any external source.
  control_plane_ip_filter = ["0.0.0.0/0"]
  name                    = "exampleapp"
  network                 = upcloud_network.example.id
  zone                    = "fi-hel2"
}

# Create a Kubernetes cluster node group
resource "upcloud_kubernetes_node_group" "group" {
  cluster    = resource.upcloud_kubernetes_cluster.example.id
  node_count = 2
  name       = "medium"
  plan       = "PREMIUM-2xCPU-4GB"

  labels = {
    managedBy = "terraform"
  }

  taint {
    effect = "NoExecute"
    key    = "taintKey"
    value  = "taintValue"
  }
}

# Create a Kubernetes cluster node group with a GPU plan, with a custom storage size
resource "upcloud_kubernetes_node_group" "group_gpu" {
  cluster    = resource.upcloud_kubernetes_cluster.example.id
  node_count = 2
  name       = "gpu-workers"
  plan       = "GPU-8xCPU-64GB-1xL40S"
  gpu_plan {
    storage_size = 250
  }
  labels = {
    gpu = "NVIDIA-L40S"
  }
}

# Create a Kubernetes cluster node group with a Cloud Native plan, with a custom storage size and tier
resource "upcloud_kubernetes_node_group" "group_cloud_native" {
  cluster    = resource.upcloud_kubernetes_cluster.example.id
  node_count = 4
  name       = "cloud-native-workers"
  plan       = "CLOUDNATIVE-4xCPU-8GB"
  cloud_native_plan {
    storage_size = 100
    storage_tier = "standard"
  }
}
