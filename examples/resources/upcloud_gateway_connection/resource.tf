resource "upcloud_router" "this" {
  name = "gateway-example-router"

  # UpCloud Network Gateway Service will add a static route to this router to ensure gateway networking is working as intended.
  # You need to ignore changes to it, otherwise TF will attempt to remove the static routes on subsequent applies
  lifecycle {
    ignore_changes = [static_route]
  }
}

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
  name = "gateway-example-gw"
  zone = "pl-waw1"

  # Note that VPN feature is currently in beta phase.
  # Also not all VPN plans allow VPN feature.
  # For more info see https://upcloud.com/resources/docs/networking#nat-and-vpn-gateways
  features = ["vpn"]
  plan     = "advanced"

  router {
    id = upcloud_router.this.id
  }
}

resource "upcloud_gateway_connection" "this" {
  gateway = upcloud_gateway.this.id
  name    = "test-connection"
  type    = "ipsec"

  local_route {
    name           = "local-route"
    type           = "static"
    static_network = "10.123.123.0/24"
  }

  remote_route {
    name           = "remote-route"
    type           = "static"
    static_network = "100.123.123.0/24"
  }
}
