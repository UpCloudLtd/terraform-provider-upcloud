# set the provider version
terraform {
  required_providers {
    upcloud = {
      source  = "UpCloudLtd/upcloud"
      version = "~> 2.0"
    }
  }
}

# configure the UpCloud provider
provider "upcloud" {
  # username and password configuration arguments can be omitted
  # if environment variables UPCLOUD_USERNAME and UPCLOUD_PASSWORD are set
  # username = ""
  # password = ""
}

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
