variable "prefix" {
  default = "tf-acc-test-server-metadata-"
  type    = string
}

variable "zone" {
  default = "fi-hel2"
  type    = string
}

resource "upcloud_storage" "this" {
  title = "${var.prefix}storage"
  zone  = var.zone
  size  = 10
  tier  = "maxiops"
}

resource "upcloud_server" "this" {
  hostname = "${var.prefix}server"
  zone     = var.zone
  metadata = false

  storage_devices {
    storage = upcloud_storage.this.id
  }

  network_interface {
    type = "utility"
  }
}
