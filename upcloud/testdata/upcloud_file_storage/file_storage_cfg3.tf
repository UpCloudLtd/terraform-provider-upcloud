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
    name              = "${var.prefix}${var.file-storage-name}-3"
    size              = 250
    zone              = "fi-hel2"
    configured_status = "started"

    labels = {
        single = "onlyone"
    }
}