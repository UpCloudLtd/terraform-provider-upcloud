variable "prefix" {
  default = "tf-acc-test-objsto2-"
  type    = string
}

variable "zone" {
  default = "fi-hel1"
  type    = string
}

variable "region" {
  default = "europe-1"
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
  name              = "tf-acc-test-objstov2-complex"
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
  name              = "tf-acc-test-objstov2-renamed"
  region            = var.region
  configured_status = "started"

  # No networks or labels
}
