# Build server with your latest custom image 
#
# Note that when applied new updated image will cause replacement of the old server (debian.example.tld) with the new server created based on the updated image.
# This can cause posible data loss if it hasn't been taken into account when planning the service.
data "upcloud_storage" "app_image" {
  type        = "template"
  name_regex  = "^app_image.*"
  most_recent = true
}

resource "upcloud_server" "example" {
  hostname = "debian.example.tld"
  zone     = "fi-hel1"

  network_interface {
    type = "public"
  }

  template {
    storage = data.upcloud_storage.app_image.id
  }
}
