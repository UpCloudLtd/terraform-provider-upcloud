# set the provider version
terraform {
  required_providers {
    upcloud = {
      source = "UpCloudLtd/upcloud"
      version = "~> 2.0"
    }
  }
}

# configure the UpCloud provider
provider "upcloud" {}

# create SDN private network with DHCP enabled on 10.0.0.0/24
resource "upcloud_network" "example_network" {
  name = "Example Private Network"
  zone = "de-fra1"

  ip_network {
    address = "10.0.0.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

# create a server attached to the network
resource "upcloud_server" "example" {
  hostname = "ubuntu.example.tld"
  zone     = "de-fra1"

  network_interface {
    type = "public"
  }

  network_interface {
    type = "private"
    network = upcloud_network.example_network.id
  }

  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
    size = 10
  }
}
