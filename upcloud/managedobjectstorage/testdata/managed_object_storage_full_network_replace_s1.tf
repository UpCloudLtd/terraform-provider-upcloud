variable "prefix" {
  default = "tf-acc-test-objsto-fullswap-"
  type    = string
}

variable "zone" {
  default = "se-sto1"
  type    = string
}

variable "region" {
  default = "europe-3"
  type    = string
}

variable "cidr_a" {
  type = string
}

variable "cidr_b" {
  type = string
}

resource "upcloud_router" "this" {
  name = "${var.prefix}router"
}

resource "upcloud_network" "private_a" {
  name = "${var.prefix}private-a"
  zone = var.zone

  ip_network {
    address = var.cidr_a
    dhcp    = false
    family  = "IPv4"
  }

  router = upcloud_router.this.id
}

resource "upcloud_managed_object_storage" "this" {
  name              = "${var.prefix}service"
  region            = var.region
  configured_status = "started"

  network {
    family = "IPv4"
    name   = "public"
    type   = "public"
  }

  network {
    family = "IPv4"
    name   = "private"
    type   = "private"
    uuid   = upcloud_network.private_a.id
  }
}
