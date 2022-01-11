terraform {
  required_providers {
    upcloud = {
      source  = "UpCloudLtd/upcloud"
      version = "~> 2.0"
    }
  }
}

provider "upcloud" {
  # username and password configuration arguments can be omitted  
  # if environment variables UPCLOUD_USERNAME and UPCLOUD_PASSWORD are set
  # username = ""
  # password = ""
}
