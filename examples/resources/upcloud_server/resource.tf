resource "upcloud_server" "example" {
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"

  # The metadata service must be enabled when using recent cloud-init based templates.
  metadata = true

  template {
    storage = "Ubuntu Server 24.04 LTS (Noble Numbat)"
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
