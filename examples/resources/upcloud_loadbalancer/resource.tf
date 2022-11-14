variable "lb_zone" {
  type    = string
  default = "fi-hel2"
}

resource "upcloud_network" "lb_network" {
  name = "lb-test-net"
  zone = var.lb_zone
  ip_network {
    address = "10.0.0.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_loadbalancer" "lb" {
  configured_status = "started"
  name              = "lb-test"
  plan              = "development"
  zone              = var.lb_zone
  networks {
    name    = "Private-Net"
    type    = "private"
    family  = "IPv4"
    network = resource.upcloud_network.lb_network.id
  }
  networks {
    name   = "Public-Net"
    type   = "public"
    family = "IPv4"
  }
}
