# The following example defines a server and then links the server to a single firewall rule. 
# The list of firewall rules applied to the server can be expanded by providing additional server_firewall_rules blocks.

resource "upcloud_server" "example" {
  firewall = true
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"

  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
    size    = 25
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
