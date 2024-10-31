variable "prefix" {
  default = "tf-acc-test-server-interfaces-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

resource "upcloud_network" "this" {
  name = "${var.prefix}net"
  zone = var.zone

  ip_network {
    address = "10.100.3.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_floating_ip_address" "this" {
  mac_address = upcloud_server.this.network_interface[0].mac_address
}

resource "upcloud_server" "this" {
  hostname = "${var.prefix}vm"
  title    = "${var.prefix}vm"
  zone     = var.zone
  plan     = "1xCPU-1GB"
  metadata = true

  login {
    keys = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIB8Q"]
  }

  // Remove interface with index 2

  network_interface {
    type  = "public"
    index = 1
  }

  network_interface {
    type    = "private"
    index   = 3
    network = upcloud_network.this.id
  }

  network_interface {
    type    = "private"
    index   = 4
    ip_address = "10.100.3.30"
    network = upcloud_network.this.id
  }

  network_interface {
    index = 5
    type  = "utility"
  }

  template {
    title   = "${var.prefix}disk"
    storage = "Ubuntu Server 22.04 LTS (Jammy Jellyfish)"
    size    = 20
  }
}
