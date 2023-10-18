// Create router for the gateway
resource "upcloud_router" "this" {
  name = "gateway-example-router"

  # UpCloud Network Gateway Service will add a static route to this router to ensure gateway networking is working as intended.
  # You need to ignore changes to it, otherwise TF will attempt to remove the static routes on subsequent applies
  lifecycle {
    ignore_changes = [static_route]
  }
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
