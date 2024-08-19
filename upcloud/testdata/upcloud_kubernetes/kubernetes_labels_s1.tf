variable "basename" {
  default = "tf-acc-test-k8s-labels-"
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
    address = "172.23.8.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_kubernetes_cluster" "main" {
  control_plane_ip_filter = ["0.0.0.0/0"]
  name                    = "${var.basename}cluster"
  network                 = upcloud_network.main.id
  zone                    = var.zone

  labels = {
    test = "terraform-provider-acceptance-test"
  }
  storage_encryption = "data-at-rest"
}
