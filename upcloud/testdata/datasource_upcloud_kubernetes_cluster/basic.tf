variable "name" {
  default = "acc-test-datasource-upcloud-kubernetes-cluster-basic"
  type    = string
}

variable "zone" {
  default = "de-fra1"
  type    = string
}

resource "upcloud_network" "basic" {
  name = var.name
  zone = var.zone
  ip_network {
    address = "172.16.1.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_kubernetes_cluster" "basic" {
  name    = "var.name"
  network = resource.upcloud_network.cluster_private_network.id
  node_groups = [
    {
      count = 1
      name  = var.name
      plan  = "K8S-2xCPU-4GB"
    }
  ]
  zone = var.zone
}

data "upcloud_kubernetes_cluster" "basic" {
  id = resource.upcloud_kubernetes_cluster.basic.id
}
