// Create network for the File Storage
resource "upcloud_network" "this" {
  name = "file-storage-net-test"
  zone = "fi-hel2"

  ip_network {
    address = "172.16.8.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_file_storage" "example" {
  name              = "example-file-storage-test"
  size              = 250
  zone              = "fi-hel2"
  configured_status = "stopped"

  labels = {
    environment = "staging"
    customer    = "example-customer"
  }

  network {
    family     = "IPv4"
    name       = "example-private-net"
    uuid       = upcloud_network.this.id
    ip_address = "172.16.8.11"
  }
}

resource "upcloud_file_storage_share" "example" {
  file_storage = upcloud_file_storage.example.id
  name         = "write-to-project"
  path         = "/project"
}

resource "upcloud_file_storage_share_acl" "example" {
  file_storage = upcloud_file_storage.example.id
  share_name   = upcloud_file_storage_share.example.name
  name         = "acl-for-project"
  target       = "172.16.8.12"
  permission   = "rw"
}
