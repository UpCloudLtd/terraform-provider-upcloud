variable "prefix" {
  default = "tf-acc-test-file-storage-"
  type    = string
}

variable "net-name" {
  default = "net-name"
  type    = string
}

variable "file-storage-name" {
  default = "file-storage-name"
  type    = string
}

variable "cidr" {
  default = "172.16.34.0/24"
  type    = string
}

variable "network-ip-addrs" {
  default = "172.16.34.57"
  type    = string
}

resource "upcloud_network" "this" {
    name = "${var.prefix}${var.net-name}"
    zone = "fi-hel2"

    ip_network {
        address = "${var.cidr}"
        dhcp    = true
        family  = "IPv4"
    }
}

resource "upcloud_file_storage" "example" {
    name              = "${var.prefix}${var.file-storage-name}_v4"
    size              = 250
    zone              = "fi-hel2"
    configured_status = "stopped"

    labels = {
        single = "onlyone"
    }

  network {
    family     = "IPv4"
    name       = "example-private-net-readd"
    uuid       = upcloud_network.this.id
    ip_address = "${var.network-ip-addrs}"
  }
}