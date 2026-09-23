# Attach an SDN private firewall ruleset to a server
resource "upcloud_network" "example" {
  name = "terraform-example-private-network"
  zone = "de-fra1"

  ip_network {
    address = "172.24.1.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_server" "example" {
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"

  template {
    storage = "Ubuntu Server 24.04 LTS (Noble Numbat)"
  }

  network_interface {
    type    = "private"
    network = upcloud_network.example.id
  }
}

resource "upcloud_firewall_ruleset" "example" {
  name        = "example-private-ruleset"
  description = "Private SDN ruleset"

  rules = [
    {
      action                 = "accept"
      direction              = "in"
      family                 = "IPv4"
      protocol               = "tcp"
      comment                = "Allow SSH"
      destination_port_start = 22
      destination_port_end   = 22
    }
  ]
}

resource "upcloud_server_private_firewall_ruleset" "example" {
  server_id  = upcloud_server.example.id
  ruleset_id = upcloud_firewall_ruleset.example.id
}
