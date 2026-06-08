variable "prefix" {
  default = "tf-acc-test-router-detach-"
  type    = string
}

variable "zone" {
  default = "fi-hel1"
  type    = string
}

resource "upcloud_router" "this" {
  name = "${var.prefix}router"
}

resource "upcloud_network" "this" {
  name = "${var.prefix}net"
  zone = var.zone

  # router = upcloud_router.this.id 

  ip_network {
    address = "10.0.2.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}
