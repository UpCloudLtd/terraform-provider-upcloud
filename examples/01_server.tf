provider "upcloud" {
  # You need to set UpCloud credentials in shell environment variable  # using .bashrc, .zshrc or similar  # export UPCLOUD_USERNAME="Username for Upcloud API user"  # export UPCLOUD_PASSWORD="Password for Upcloud API user"
  username = "username"
  password = "password"
}

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
}
