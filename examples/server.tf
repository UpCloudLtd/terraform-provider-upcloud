# set the provider version
terraform {
  required_providers {
    upcloud = {
      source  = "UpCloudLtd/upcloud"
      version = "~> 2.0"
    }
  }
}

provider "upcloud" {
  # Your UpCloud credentials are read from the environment variables:
  # export UPCLOUD_USERNAME="Username of your UpCloud API user"
  # export UPCLOUD_PASSWORD="Password of your UpCloud API user"
}

# create a server
resource "upcloud_server" "ubuntu" {
  hostname = "ubuntu.example.tld"
  zone     = "de-fra1"
  firewall = false

  # choose a simple plan
  plan = "1xCPU-1GB"

  # use flexible plan:
  # cpu = "2"
  # mem = "1024"

  # add a public IP address
  network_interface {
    type = "public"
  }

  # add the utility network
  network_interface {
    type = "utility"
  }

  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"

    # use size allotted by the 1xCPU-1GB plan:
    size = 25

    # UUID also works:
    # storage = "01000000-0000-4000-8000-000030200200"

    backup_rule {
      interval  = "daily"
      time      = "0100"
      retention = 8
    }
  }

  # attach an extra storage device (configured below)
  storage_devices {
    storage = upcloud_storage.datastorage.id

    # defaults if not specified:
    # address = "virtio"
    # type    = "disk"
  }

  # login details
  login {
    # create a new sudo user called "terraform"
    user = "terraform"

    keys = [
      "<YOUR PUBLIC SSH KEY HERE>",
    ]

    # create a password (set to false if using SSH key only)
    create_password = true

    # remove password_delivery if using SSH key only
    password_delivery = "sms"
  }

  # Allow terraform to connect to the server
  connection {
    # the server public IP address
    host = self.network_interface[0].ip_address
    type = "ssh"

    # the user created above
    user        = "terraform"
    private_key = "<PATH TO YOUR PRIVATE SSH KEY>"
  }
}

# create a storage for data
resource "upcloud_storage" "datastorage" {
  title = "/data"
  size  = 10
  # zone needs to match server's zone:
  zone = "de-fra1"
  tier = "maxiops"

  # backup_rule {
  #   interval  = "daily"
  #   time      = "0100"
  #   retention = 8
  # }
}

output "Public_ip" {
  value = upcloud_server.ubuntu.network_interface[0].ip_address
}
