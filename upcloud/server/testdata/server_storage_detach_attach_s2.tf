variable "zone" {
  default = "fi-hel1"
  type    = string
}

resource "upcloud_storage" "shared" {
  title = "tf-acc-test-server-storage-detach-attach-disk"
  size  = 10
  zone  = var.zone
}

resource "upcloud_server" "server_a" {
  hostname = "tf-acc-test-server-storage-detach-attach-a"
  zone     = var.zone
  plan     = "1xCPU-1GB"
  metadata = true

  template {
    storage = "01000000-0000-4000-8000-000020070100"
    size    = 25
  }

  network_interface {
    type = "utility"
  }
}

resource "upcloud_server" "server_b" {
  hostname = "tf-acc-test-server-storage-detach-attach-b"
  zone     = var.zone
  plan     = "1xCPU-1GB"
  metadata = true

  template {
    storage = "01000000-0000-4000-8000-000020070100"
    size    = 25
  }

  network_interface {
    type = "utility"
  }

  storage_devices {
    storage = upcloud_storage.shared.id
  }
}
