# set the provider version
terraform {
  required_providers {
    upcloud = {
      source = "UpCloudLtd/upcloud"
      version = "~> 2.0"
    }
  }
}

# configure the provider
provider "upcloud" {
  # Your UpCloud credentials are read from the environment variables:
  # export UPCLOUD_USERNAME="Username of your UpCloud API user"
  # export UPCLOUD_PASSWORD="Password of your UpCloud API user"
}

# create a server
resource "upcloud_server" "example" {
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"

  # Declare network interfaces
  network_interface {
    type = "public"
  }
  network_interface {
    type = "utility"
  }

  # Include at least one public SSH key
  login {
    user = "terraform"
    keys = [
      "<YOUR SSH PUBLIC KEY>",
    ]
    create_password = false
  }

  # Provision the server with Ubuntu
  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"

    # Use all the space allotted by the selected simple plan
    size = 25

    # Enable backups
    backup_rule {
      interval  = "daily"
      time      = "0100"
      retention = 8
    }
  }

  # Configuring connection details
  connection {
    # The server public IP address
    host        = self.network_interface[0].ip_address
    type        = "ssh"
    user        = "terraform"
    private_key = "<PATH TO YOUR SSH PRIVATE KEY>"
  }

  # Remotely executing a command on the server
  provisioner "remote-exec" {
    inline = [
      "echo 'Hello world!' > /tmp/jeejee"
    ]
  }
}
