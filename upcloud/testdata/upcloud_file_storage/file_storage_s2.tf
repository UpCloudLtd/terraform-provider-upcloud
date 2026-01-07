variable "prefix" {
  default = "tf-acc-test-file-storage-"
  type    = string
}

variable "suffix" {
  default = "suffix"
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

variable "zone" {
  default = "fi-hel2"
  type    = string
}

resource "upcloud_network" "this" {
  name = "${var.prefix}${var.suffix}"
  zone = var.zone

  ip_network {
    address = var.cidr
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_file_storage" "this" {
  name              = "${var.prefix}${var.suffix}-s2"
  size              = 250
  zone              = var.zone
  configured_status = "started"

  labels = {
    environment = "staging"
    customer    = "example-customer"
    env         = "test"
  }
}

resource "upcloud_file_storage_share" "this" {
  file_storage = upcloud_file_storage.this.id
  name         = "${var.prefix}${var.suffix}-s2"
  path         = "/project"
}

resource "upcloud_file_storage_share_acl" "this" {
  file_storage = upcloud_file_storage.this.id
  share_name   = upcloud_file_storage_share.this.name
  name         = "${var.prefix}${var.suffix}"
  target       = var.acl-1-ip
  permission   = "rw"
}

resource "upcloud_file_storage_share" "this2" {
  file_storage = upcloud_file_storage.this.id
  name         = "${var.prefix}${var.suffix}-2-s2"
  path         = "/project2"
}

resource "upcloud_file_storage_share_acl" "this2" {
  file_storage = upcloud_file_storage.this.id
  share_name   = upcloud_file_storage_share.this2.name
  name         = "${var.prefix}${var.suffix}-2"
  target       = var.acl-2-ip
  permission   = "ro"
}
