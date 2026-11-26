locals {
  name_prefix = "${var.basename}k8s-oidc-e2e-"
}

resource "upcloud_router" "main" {
  name = "${local.name_prefix}router"
}

resource "upcloud_network" "main" {
  name   = "${local.name_prefix}net"
  zone   = var.zone
  router = upcloud_router.main.id

  ip_network {
    address            = var.network_cidr
    dhcp               = true
    dhcp_default_route = false
    family             = "IPv4"
  }
}

resource "upcloud_kubernetes_cluster" "main" {
  control_plane_ip_filter = ["0.0.0.0/0"]
  name                    = "${local.name_prefix}cluster"
  network                 = upcloud_network.main.id
  zone                    = var.zone
}

resource "upcloud_kubernetes_node_group" "default" {
  cluster       = upcloud_kubernetes_cluster.main.id
  node_count    = 1
  anti_affinity = true
  name          = "default"
  plan          = "STARTER-2xCPU-4GB"
}

ephemeral "upcloud_kubernetes_kubeconfig" "main" {
  cluster = upcloud_kubernetes_cluster.main.id
}
