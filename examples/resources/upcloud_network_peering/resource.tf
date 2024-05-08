# Network peering requires the networks to have routers attached to them.
resource "upcloud_router" "this" {
  name = "network-peering-example-router"
}

resource "upcloud_network" "example" {
  name   = "network-peering-example-net"
  zone   = "nl-ams1"
  router = upcloud_router.example.id

  ip_network {
    address = "10.0.0.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_network_peering" "this" {
  count = 1
  name  = "network-peering-example-peering"

  network {
    uuid = upcloud_network.example.id
  }

  peer_network {
    uuid = "0305723a-e5cb-4ef6-985d-e36ed44d133a"
  }
}
