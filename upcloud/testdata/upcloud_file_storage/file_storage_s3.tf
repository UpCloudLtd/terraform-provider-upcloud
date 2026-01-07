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

variable "zone" {
  default = "fi-hel2"
  type    = string
}

resource "upcloud_network" "this" {
    name = "${var.prefix}${var.suffix}"
    zone = var.zone

    ip_network {
        address = "${var.cidr}"
        dhcp    = true
        family  = "IPv4"
    }
}
resource "upcloud_file_storage" "this" {
    name              = "${var.prefix}${var.suffix}-s3"
    size              = 250
    zone              = var.zone
    configured_status = "started"

    labels = {
        single = "onlyone"
    }
}