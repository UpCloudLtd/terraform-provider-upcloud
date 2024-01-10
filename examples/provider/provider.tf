terraform {
  required_providers {
    upcloud = {
      source  = "UpCloudLtd/upcloud"
      version = "~> 3.0"
    }
  }
}

provider "upcloud" {
  username = "<Your username>"
  password = "<Your password>"
}

