terraform {
  required_providers {
    upcloud = {
      source  = "UpCloudLtd/upcloud"
      version = "~> 4.0"
    }
  }
}

provider "upcloud" {
  username = "<Your username>"
  password = "<Your password>"
}

