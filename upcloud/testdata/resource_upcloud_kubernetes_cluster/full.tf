variable "name" {
  default = "acc-test-resource-upcloud-kubernetes-cluster-full"
  type    = string
}

variable "zone" {
  default = "de-fra1"
  type    = string
}

resource "upcloud_network" "full" {
  name = "terraform-provider-upcloud-test"
  zone = var.zone
  ip_network {
    address = "172.16.1.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_kubernetes_cluster" "full" {
  name    = var.name
  network = resource.upcloud_network.cluster_private_network.id
  node_groups = [
    {
      count = 1
      labels = {
        env       = "dev"
        managedBy = var.name
      }
      name = "small"
      plan = "K8S-2xCPU-4GB"
    },
    {
      count = 1
      labels = {
        env       = "qa"
        managedBy = var.name
      }
      name = "medium"
      plan = "K8S-4xCPU-8GB"
    }
  ]
  storage = "01000000-0000-4000-8000-000160010100"
  zone    = var.zone
}
