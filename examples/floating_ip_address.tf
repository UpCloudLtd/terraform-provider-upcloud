# set the provider version
terraform {
  required_providers {
    upcloud = {
      source = "UpCloudLtd/upcloud"
      version = "~> 2.0"
    }
  }
}

# configure the provider
provider "upcloud" {
  # Your UpCloud credentials are read from the environment variables:
  # export UPCLOUD_USERNAME="Username of your UpCloud API user"
  # export UPCLOUD_PASSWORD="Password of your UpCloud API user"
}

# create a server
resource "upcloud_server" "loadbalancer" {
  hostname = "lb.example.tld"
  zone     = "nl-ams1"
  plan     = "1xCPU-1GB"

  # Declare a network interface for the floating IP
  network_interface {
    type = "public"
  }

  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
    size    = 25
  }

  login {
    user = "terraform"
    keys = [
      "<YOUR PUBLIC SSH KEY HERE>",
    ]
    create_password = false
  }
}

# create a floating ip address
resource "upcloud_floating_ip_address" "my_floating_ip" {
  # attach the floating IP address to the server's public interface MAC
  mac_address = upcloud_server.loadbalancer.network_interface[0].mac_address
  zone        = "de-fra1"
}

output "Public_ip" {
  value = upcloud_server.loadbalancer.network_interface[0].ip_address
}
