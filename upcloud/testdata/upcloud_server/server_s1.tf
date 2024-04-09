variable "prefix" {
  default = "tf-acc-test-server-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

resource "upcloud_network" "n1" {
  name = "${var.prefix}n1"
  zone = var.zone
  ip_network {
    address = "172.102.0.0/16"
    dhcp    = false
    family  = "IPv4"
  }
}

resource "upcloud_server" "server1" {
  hostname = "${var.prefix}server1"
  zone     =  var.zone

  network_interface {
    type = "private"
    network = upcloud_network.n1.id
    ip_address = "172.102.0.2"
    additional_ip_address {
      ip_address = "172.102.0.3"
    }
  }

  login {
    keys = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIB8Q"]
  }

  template {
    storage = "Debian GNU/Linux 12 (Bookworm)"
  }

  metadata = true
  firewall = true
}