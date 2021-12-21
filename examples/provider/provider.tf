terraform {
  required_version = ">= 0.12.0"
}

provider "upcloud" {
  username = "<Your username>"
  password = "<Your password>"
}

resource "upcloud_server" "myserver" {
  # ...
}
