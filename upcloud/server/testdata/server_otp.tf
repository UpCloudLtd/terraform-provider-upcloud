variable "prefix" {
  default = "tf-acc-test-server-otp-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

variable "step" {
  default = ""
  type    = string
}

resource "upcloud_server" "this" {
  hostname = "${var.prefix}${var.step}vm"
  zone     = var.zone
  plan     = "1xCPU-1GB"

  template {
    storage = "AlmaLinux 8"
  }

  network_interface {
    type = "public"
  }

  login {
    create_password   = true
    password_delivery = "none"
  }
}
