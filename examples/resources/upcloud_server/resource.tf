resource "upcloud_server" "example" {
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"

  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
    size    = 25

    backup_rule {
      interval  = "daily"
      time      = "0100"
      retention = 8
    }
  }

  network_interface {
    type = "public"
  }

  labels = {
    env        = "dev"
    production = "false"
  }

  login {
    user = "myusername"

    keys = [
      "<YOUR SSH PUBLIC KEY>",
    ]
  }
}
