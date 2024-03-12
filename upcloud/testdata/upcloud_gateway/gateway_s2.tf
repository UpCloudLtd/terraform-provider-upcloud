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
  features          = ["nat", "vpn"]
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

resource "upcloud_gateway_connection" "this" {
  gateway = upcloud_gateway.this.id
  name    = "test-connection"
  type    = "ipsec"

  local_route {
    name           = "local-route-updated"
    type           = "static"
    static_network = "11.123.123.0/24"
  }

  remote_route {
    name           = "remote-route-updated"
    type           = "static"
    static_network = "111.123.123.0/24"
  }
}
