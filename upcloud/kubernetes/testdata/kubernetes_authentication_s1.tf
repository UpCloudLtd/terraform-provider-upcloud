variable "basename" {
  default = "tf-acc-test-k8s-auth-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

resource "upcloud_router" "main" {
  name = "${var.basename}router"
}

resource "upcloud_network" "main" {
  name   = "${var.basename}net"
  zone   = var.zone
  router = upcloud_router.main.id

  ip_network {
    address = "172.23.9.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_kubernetes_cluster" "main" {
  control_plane_ip_filter = ["0.0.0.0/0"]
  name                    = "${var.basename}cluster"
  network                 = upcloud_network.main.id
  zone                    = var.zone
}

resource "upcloud_kubernetes_cluster_authentication" "main" {
  cluster    = upcloud_kubernetes_cluster.main.id
  name       = "example"
  issuer_url = "https://accounts.example.com"
  audiences  = ["client-id-123"]
}
