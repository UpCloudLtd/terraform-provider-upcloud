variable "prefix" {
  default = "tf-acc-test-server-tags-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

variable "tag_suffix" {
  default = ""
  type    = string
}

resource "upcloud_tag" "this" {
  name = "${var.prefix}${var.tag_suffix}"
}

resource "upcloud_server" "this" {
  hostname = "${var.prefix}vm"
  zone     = var.zone
  plan     = "1xCPU-1GB"
  metadata = true

  template {
    storage = "Ubuntu Server 24.04 LTS (Noble Numbat)"
  }

  login { password_delivery = "none" }

  network_interface {
    type = "public"
  }

  tags = [upcloud_tag.this.name]
}
