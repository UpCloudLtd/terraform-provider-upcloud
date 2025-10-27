# The following example defines a server and then links the server to a single firewall rule. 
# The list of firewall rules applied to the server can be expanded by providing additional server_firewall_rules blocks.

resource "upcloud_server" "example" {
  firewall = true
  hostname = "terraform.example.tld1"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"
  metadata = true

  login {
    password_delivery = "none"
  }

  template {
    storage = "Ubuntu Server 24.04 LTS (Noble Numbat)"
  }

  network_interface {
    type = "utility"
  }
}

resource "upcloud_firewall_rules" "example" {
  server_id = upcloud_server.example.id

  firewall_rule {
    action                 = "accept"
    comment                = "Allow SSH from this network"
    destination_port_end   = "22"
    destination_port_start = "22"
    direction              = "in"
    family                 = "IPv4"
    protocol               = "tcp"
    source_address_end     = "192.168.1.255"
    source_address_start   = "192.168.1.1"
  }
}
