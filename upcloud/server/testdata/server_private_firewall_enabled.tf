variable "name" {
  type = string
}

variable "template" {
  type = string
}

resource "upcloud_network" "test" {
  name = var.name
  zone = "fi-hel1"

  ip_network {
    address = "172.24.1.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

# Separate storage keeps import verification independent of template clone provenance.
resource "upcloud_storage" "test" {
  title = var.name
  zone  = "fi-hel1"
  size  = 10
  tier  = "maxiops"

  clone {
    id = var.template
  }
}

resource "upcloud_server" "test" {
  hostname                                 = var.name
  zone                                     = "fi-hel1"
  plan                                     = "1xCPU-1GB"
  metadata                                 = true
  firewall                                 = true
  firewall_private                         = true
  firewall_private_default_incoming_action = "drop"
  firewall_private_default_outgoing_action = "accept"

  storage_devices {
    storage = upcloud_storage.test.id
  }

  network_interface {
    type    = "private"
    network = upcloud_network.test.id
  }
}

data "upcloud_server" "test" {
  id = upcloud_server.test.id
}
