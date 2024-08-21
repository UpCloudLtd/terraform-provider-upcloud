variable "basename" {
  default = "tf-acc-test-k8s-storage-encryption-"
  type    = string
}

variable "zone" {
  default = "es-mad1"
  type    = string
}

resource "upcloud_router" "main" {
  name = "${var.basename}router"
}

resource "upcloud_network" "main" {
  name   = "${var.basename}network"
  zone   = var.zone
  router = upcloud_router.main.id

  ip_network {
    address = "172.23.9.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_kubernetes_cluster" "main" {
  name                    = "${var.basename}cluster"
  network                 = upcloud_network.main.id
  storage_encryption      = "data-at-rest"
  zone                    = var.zone
  control_plane_ip_filter = ["0.0.0.0/0"]
}

resource "upcloud_kubernetes_node_group" "main" {
  cluster    = resource.upcloud_kubernetes_cluster.main.id
  node_count = 1
  name       = "small"
  plan       = "1xCPU-1GB"
}
