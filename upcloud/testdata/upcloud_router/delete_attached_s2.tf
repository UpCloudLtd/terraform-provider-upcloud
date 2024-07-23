variable "prefix" {
  default = "tf-acc-test-router-delete-attached-"
  type    = string
}

variable "zone" {
  default = "fi-hel1"
  type    = string
}

# Delete the router

resource "upcloud_network" "this" {
  name = "${var.prefix}net"
  zone = var.zone

  ip_network {
    address = "10.0.3.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}
