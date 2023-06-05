// Create router for the gateway
resource "upcloud_router" "this" {
  name = "gateway-example-router"
}

// Create network for the gateway
resource "upcloud_network" "this" {
  name = "gateway-example-net"
  zone = "pl-waw1"

  ip_network {
    address = "172.16.2.0/24"
    dhcp    = true
    family  = "IPv4"
  }

  router = upcloud_router.this.id
}

resource "upcloud_gateway" "this" {
  name     = "gateway-example-gw"
  zone     = "pl-waw1"
  features = ["nat"]

  router {
    id = upcloud_router.this.id
  }

  labels = {
    managed-by = "terraform"
  }
}
