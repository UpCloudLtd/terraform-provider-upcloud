variable "prefix" {
  default = "tf-acc-test-server-server-group-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

resource "upcloud_server" "test_1" {
  zone     = var.zone
  hostname = "${var.prefix}-vm-1"
  plan     = "1xCPU-1GB"

  template {
    storage = "Debian GNU/Linux 11 (Bullseye)"
    size    = 25
  }

  network_interface {
    type = "public"
  }
}

resource "upcloud_server" "test_2" {
  zone         = var.zone
  hostname     = "${var.prefix}-vm-2"
  plan         = "1xCPU-1GB"
  server_group = upcloud_server_group.test.id

  template {
    storage = "Debian GNU/Linux 11 (Bullseye)"
    size    = 25
  }

  network_interface {
    type = "public"
  }
}

resource "upcloud_server_group" "test" {
  title                = "${var.prefix}-vmgroup"
  anti_affinity_policy = "no"
  track_members        = false
}
