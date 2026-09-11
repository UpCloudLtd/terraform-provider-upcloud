variable "basename" {
  default = "tf-acc-test-k8s-storage-encryption-"
  type    = string
}

variable "zone1" {
  default = "es-mad1"
  type    = string
}

variable "zone2" {
  default = "fi-hel2"
  type    = string
}

resource "upcloud_router" "data-at-rest" {
  name = "${var.basename}router"
}

resource "upcloud_network" "data-at-rest" {
  name   = "${var.basename}network"
  zone   = var.zone1
  router = upcloud_router.data-at-rest.id

  ip_network {
    address = "172.23.50.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_kubernetes_cluster" "data-at-rest" {
  name                    = "${var.basename}cluster"
  network                 = upcloud_network.data-at-rest.id
  storage_encryption      = "data-at-rest"
  zone                    = var.zone1
  control_plane_ip_filter = ["0.0.0.0/0"]
}

resource "upcloud_kubernetes_node_group" "data-at-rest" {
  cluster    = resource.upcloud_kubernetes_cluster.data-at-rest.id
  node_count = 1
  name       = "small"
  plan       = "1xCPU-2GB"
}

resource "upcloud_router" "none" {
  name = "${var.basename}router2"
}

resource "upcloud_network" "none" {
  name   = "${var.basename}network2"
  zone   = var.zone2
  router = upcloud_router.none.id

  ip_network {
    address = "172.23.60.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_kubernetes_cluster" "none" {
  name                    = "${var.basename}cluster2"
  network                 = upcloud_network.none.id
  zone                    = var.zone2
  control_plane_ip_filter = ["0.0.0.0/0"]
}

resource "upcloud_kubernetes_node_group" "none" {
  cluster    = resource.upcloud_kubernetes_cluster.none.id
  node_count = 1
  name       = "small"
  plan       = "1xCPU-2GB"

  // This should not cause inconsistency in the plan as the node group was created with default storage_encryption setting that in practice matches the value defined here.
  storage_encryption = "none"
}
