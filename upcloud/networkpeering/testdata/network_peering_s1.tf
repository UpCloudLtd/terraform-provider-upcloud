variable "prefix" {
  default = "tf-acc-test-peering-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

locals {
  cidr = ["172.18.221.0/24", "172.18.222.0/24"]
  peering_labels = [{
    test     = "tf-acc-test"
    owned-by = "team-devex"
  }, {}]
}

resource "upcloud_router" "this" {
  count = 2
  name  = "${var.prefix}router-${count.index}"
}


resource "upcloud_network" "this" {
  count = 2
  name  = "${var.prefix}net-${count.index}"
  zone  = "pl-waw1"

  ip_network {
    address = local.cidr[count.index]
    dhcp    = true
    family  = "IPv4"
  }

  router = upcloud_router.this[count.index].id
}

resource "upcloud_network_peering" "this" {
  count  = 2
  name   = "${var.prefix}peering-${count.index}-to-${(count.index + 1) % 2}"
  labels = local.peering_labels[count.index]

  network {
    uuid = upcloud_network.this[count.index].id
  }

  peer_network {
    uuid = upcloud_network.this[(count.index + 1) % 2].id
  }
}
