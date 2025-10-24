
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

variable "acl-1-ip" {
  default = "172.16.34.45"
  type    = string
}

variable "acl-2-ip" {
  default = "172.16.34.46"
  type    = string
}

variable "network-ip-addrs" {
  default = "172.16.34.50"
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
  name              = "${var.prefix}${var.file-storage-name}-2"
  size              = 250
  zone              = "fi-hel2"
  configured_status = "started"

  labels = {
    environment = "staging"
    customer    = "example-customer"
    env         = "test"
  }

  share {
    name = "write-to-project"
    path = "/project"
    acl {
      target     = "${var.acl-1-ip}"
      permission = "rw"
    }
  }

  share {
    name = "read-only"
    path = "/public"
    acl {
      target     = "${var.acl-2-ip}"
      permission = "ro"
    }
  }

  network = {
    family     = "IPv4"
    name       = "example-private-net"
    uuid       = upcloud_network.this.id
    ip_address = "${var.network-ip-addrs}"
  }
}