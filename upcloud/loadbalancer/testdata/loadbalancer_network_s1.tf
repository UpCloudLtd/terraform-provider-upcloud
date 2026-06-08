variable "basename" {
  type    = string
  default = "tf-acc-test-lb-network-migrate-"
}

variable "zone" {
  type    = string
  default = "fi-hel2"
}

resource "upcloud_network" "this" {
  name = "${var.basename}net"
  zone = var.zone

  ip_network {
    address = "10.0.8.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_loadbalancer" "migrate_then_rename" {
  name = "${var.basename}lb"
  plan = "development"
  zone = var.zone

  network = upcloud_network.this.id
}

resource "upcloud_loadbalancer" "migrate_and_rename" {
  name = "${var.basename}and-rename-lb"
  plan = "development"
  zone = var.zone

  network = upcloud_network.this.id
}
