terraform {
  required_providers {
    upcloud = {
      source  = "UpCloudLtd/upcloud"
      version = "~> 5.0"
    }
  }
}

provider "upcloud" {
  # If the UPCLOUD_TOKEN environment variable is set,
  # you can omit the token argument.
  # token = ""
  #
  # To use a subaccount's username and password instead,
  # set the UPCLOUD_USERNAME and UPCLOUD_PASSWORD environment variables.
}
