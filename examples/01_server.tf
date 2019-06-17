provider "upcloud" {
  # You need to set UpCloud credentials in shell environment variable
  # using .bashrc, .zshrc or similar
  # export UPCLOUD_USERNAME="Username for Upcloud API user"
  # export UPCLOUD_PASSWORD="Password for Upcloud API user"
}

resource "upcloud_server" "test" {
  zone     = "fi-hel1"
  hostname = "ubuntu.example.tld"

  cpu      = "2"
  mem      = "1024"

  # Login details
  login {
    user = "tf"

    keys = [
      "ssh-rsa xx",
    ]

    create_password   = true
    password_delivery = "sms"
  }

  storage_devices = [
    {
      # You can use both storage template names and UUIDs
      size    = 50
      action  = "clone"
      tier    = "maxiops"
      storage = "Ubuntu Server 16.04 LTS (Xenial Xerus)"

      backup_rule = {
        interval = "daily"
        time = "0100"
        retention = 8
      }
    }
  ]
}

resource "upcloud_tag" "My-tag" {
  name        = "TagName1"
  description = "TagDescription"
  servers     = [
     "${upcloud_server.test.id}",
     "${upcloud_server.test2.id}"
  ]
}

output "test_ipv4_address" {
  value = "${upcloud_server.test.ipv4_address}"
}

