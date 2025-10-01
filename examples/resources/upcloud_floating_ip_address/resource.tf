# Create a detached floating IP address.
resource "upcloud_floating_ip_address" "my_floating_address" {
  zone = "de-fra1"
}

# Floating IP address assigned to a server resource.
resource "upcloud_server" "example" {
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"
  metadata = true

  login {
    password_delivery = "none"
  }

  template {
    storage = "Ubuntu Server 24.04 LTS (Noble Numbat)"
  }

  network_interface {
    type = "public"
  }

}

resource "upcloud_floating_ip_address" "my_new_floating_address" {
  mac_address = upcloud_server.example.network_interface[0].mac_address
}
