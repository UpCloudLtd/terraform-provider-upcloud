variable "prefix" {
  default = "tf-acc-test-objstov2-"
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

resource "upcloud_router" "this" {
  name = "${var.prefix}router"
}

resource "upcloud_network" "this" {
  name = "${var.prefix}net"
  zone = var.zone

  ip_network {
    address = "172.18.1.0/24"
    dhcp    = false
    family  = "IPv4"
  }

  router = upcloud_router.this.id
}

resource "upcloud_managed_object_storage" "this" {
  name              = "${var.prefix}complex"
  region            = var.region
  configured_status = "started"

  network {
    family = "IPv4"
    name   = "${var.prefix}net"
    type   = "private"
    uuid   = upcloud_network.this.id
  }

  labels = {
    test     = "objsto2-tf"
    owned-by = "team-devex"
  }
}

resource "upcloud_managed_object_storage" "minimal" {
  name              = "${var.prefix}renamed"
  region            = var.region
  configured_status = "started"

  # No networks or labels
}
