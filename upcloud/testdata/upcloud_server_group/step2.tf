resource "upcloud_server" "test" {
  zone     = "pl-waw1"
  hostname = "testservergroups"
  plan     = "1xCPU-1GB"

  template {
    storage = "Debian GNU/Linux 11 (Bullseye)"
    size    = 25
  }

  network_interface {
    type = "public"
  }
}

resource "upcloud_server_group" "tf_test_1" {
  title         = "tf_test_1_updated"
  anti_affinity = true
  labels = {
    "key1" = "val1"
    "key2" = "val2"
  }
}
