provider "upcloud" {
  # You need to set UpCloud credentials in shell environment variable
  # using .bashrc, .zshrc or similar
  # export UPCLOUD_USERNAME="Username for Upcloud API user"
  # export UPCLOUD_PASSWORD="Password for Upcloud API user"
}

resource "upcloud_server" "test" {
  zone     = "fi-hel1"
  hostname = "ubuntu.example.tld"

  #cpu = "2"
  #mem = "1024"

  # Login details
  login {
    user = "tf"

    keys = [
      "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCYn6VuEgiH3//qpSa/b3Khrjy3Z4Q4fvvhRNRDrZaJqddLvQLCtoL2ktoke7+0jTcR4Vydi8bk8csUQlZxpWC6SIfif+tB8HjwusbUfLT5I5fJEI/O7gtktvtWkK4GnePFXYIdgKlXKRJ92xFnNOGV+el2zug78QahsrzsyV0Cucfjb7twPyojh5iPl3gf6f7NBHVnsqNELhJqmpo4uY+vSTfHx0siyIGP0U/Jz9dB64kbnoG6GL2fh3CEQ950Ll2luY/cfX52SO+WX/nl156A2VVCozkOSE3wbZ501Gd1508KY7ctuaqOue4DF8ZuQ1uzv4Lf9sfg4Bv4jBMTu4tvB",
    ]

    create_password   = true
    password_delivery = "sms"
  }

  storage_devices = [
    {
      # You can use both storage template names and UUIDs
      size    = 50
      action  = "clone"
      storage = "Ubuntu Server 16.04 LTS (Xenial Xerus)"

      #storage = "01ed68b5-ec65-44ec-8e98-0a9ddc187195"
    },
    {
      # Debian GNU/Linux 6.0.1 (Squeeze) (netinst)
      # Just to show you can attach different kinds of
      # resources to the instance
      action = "attach"

      storage = "01000000-0000-4000-8000-000020010301"
      type    = "cdrom"
    },
    {
      # Additional 25 GB disk
      action = "create"
      size   = 25
      tier   = "maxiops"
    },
  ]
}

output "ipv4_address" {
  value = "${upcloud_server.test.ipv4_address}"
}
