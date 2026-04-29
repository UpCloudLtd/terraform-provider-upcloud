variable "prefix" {
  default = "tf-acc-test-router-static-routes-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

resource "upcloud_router" "this" {
  name = "${var.prefix}router"

  static_route {
    name    = "test"
    route   = "192.168.240.0/24"
    nexthop = "192.168.240.10"
  }
}

resource "upcloud_network" "this" {
  name = "${var.prefix}net"
  zone = var.zone

  ip_network {
    address = "192.168.120.0/24"
    dhcp    = true
    family  = "IPv4"
  }

  router = upcloud_router.this.id
}

# The gateway is used to create service route to the routers static routes.
resource "upcloud_gateway" "this" {
  name     = "${var.prefix}gw"
  zone     = var.zone
  features = ["nat"]

  router {
    id = upcloud_router.this.id
  }
}
