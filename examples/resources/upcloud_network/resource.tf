# SDN network with a router
resource "upcloud_network" "example_network" {
  name = "example_private_net"
  zone = "nl-ams1"

  router = upcloud_router.example_router.id

  ip_network {
    address            = "10.0.0.0/24"
    dhcp               = true
    dhcp_default_route = false
    family             = "IPv4"
    gateway            = "10.0.0.1"
  }
}

resource "upcloud_router" "example_router" {
  name = "example_router"
}
