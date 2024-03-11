variable "prefix" {
  default = "tf-acc-test-net-gateway-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

resource "upcloud_router" "this" {
  name = "${var.prefix}router"

  lifecycle {
    ignore_changes = [static_route]
  }
}

resource "upcloud_network" "this" {
  name = "${var.prefix}net"
  zone = var.zone

  ip_network {
    address = "172.16.124.0/24"
    dhcp    = true
    family  = "IPv4"
  }

  router = upcloud_router.this.id
}

resource "upcloud_gateway" "this" {
  name              = "${var.prefix}gw-renamed"
  zone              = var.zone
  plan              = "advanced"
  features          = ["nat"]
  configured_status = "stopped"

  router {
    id = upcloud_router.this.id
  }

  address {
    name = "my-public-ip"
  }

  labels = {
    test     = "net-gateway-tf"
    owned-by = "team-devex"
  }
}
