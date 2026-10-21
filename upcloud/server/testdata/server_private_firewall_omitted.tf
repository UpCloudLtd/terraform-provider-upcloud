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

resource "upcloud_server" "test" {
  hostname = var.name
  zone     = "fi-hel1"
  plan     = "1xCPU-1GB"
  metadata = true
  firewall = true

  template {
    storage = var.template
    size    = 10
  }

  network_interface {
    type    = "private"
    network = upcloud_network.test.id
  }
}

data "upcloud_server" "test" {
  id = upcloud_server.test.id
}
