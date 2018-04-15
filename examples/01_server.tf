provider "upcloud" {
  # You need to set UpCloud credentials in shell environment variable  # using .bashrc, .zshrc or similar  # export UPCLOUD_USERNAME="Username for Upcloud API user"  # export UPCLOUD_PASSWORD="Password for Upcloud API user"
  username = "username"
  password = "password"
}

<<<<<<< HEAD
resource "upcloud_server" "test" {

    # System hostname
    hostname = "my-awesome-hostname"

    # Target datacenter
    zone = "fi-hel1"

    # Template name or UUID
    template = "Ubuntu Server 16.04 LTS (Xenial Xerus)"

    # Number of vCPUs
    cpu = 2

    # Amount of memory in MB
    mem = 4096

    # OS root disk size
    os_disk_size = 20

    # Login details
    login {
        user = "tf"
        keys = [
            "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCYn6VuEgiH3//qpSa/b3Khrjy3Z4Q4fvvhRNRDrZaJqddLvQLCtoL2ktoke7+0jTcR4Vydi8bk8csUQlZxpWC6SIfif+tB8HjwusbUfLT5I5fJEI/O7gtktvtWkK4GnePFXYIdgKlXKRJ92xFnNOGV+el2zug78QahsrzsyV0Cucfjb7twPyojh5iPl3gf6f7NBHVnsqNELhJqmpo4uY+vSTfHx0siyIGP0U/Jz9dB64kbnoG6GL2fh3CEQ950Ll2luY/cfX52SO+WX/nl156A2VVCozkOSE3wbZ501Gd1508KY7ctuaqOue4DF8ZuQ1uzv4Lf9sfg4Bv4jBMTu4tvB"
        ]
        create_password = true
        password_delivery = "sms"
    }
=======
resource "upcloud_server" "my-server" {
  zone     = "fi-hel1"
  hostname = "debian.example.com"

  storage_devices = [{
    size    = 50
    action  = "clone"
    storage = "01000000-0000-4000-8000-000020030100"
  },
    {
      action  = "attach"
      storage = "01000000-0000-4000-8000-000020010301"
      type    = "cdrom"
    },
    {
      action = "create"
      size   = 700
      tier   = "maxiops"
    },
  ]
>>>>>>> meafmira-master/master
}

output "ipv4_address" {
    value = "${upcloud_server.test.ipv4_address}"
}
