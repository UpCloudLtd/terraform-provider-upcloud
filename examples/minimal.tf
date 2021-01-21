# configure the UpCloud provider
provider "upcloud" {}

# create a server with minimal set of required fields
resource "upcloud_server" "example" {
  hostname = "ubuntu.example.tld"
  zone     = "de-fra1"

  network_interface {
    type = "public"
  }

  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
  }
}
