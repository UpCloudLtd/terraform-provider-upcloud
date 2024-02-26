// Create router for the network
resource "upcloud_router" "this" {
  name = "object-storage-example-router"
}

// Create network for the Managed Object Storage
resource "upcloud_network" "this" {
  name = "object-storage-example-net"
  zone = "fi-hel1"

  ip_network {
    address = "172.16.2.0/24"
    dhcp    = true
    family  = "IPv4"
  }

  router = upcloud_router.this.id
}

resource "upcloud_managed_object_storage" "this" {
  name              = "example"
  region            = "europe-1"
  configured_status = "started"

  network {
    family = "IPv4"
    name   = "example-private-net"
    type   = "private"
    uuid   = upcloud_network.this.id
  }

  labels = {
    managed-by = "terraform"
  }
}
