# Create a firewall ruleset first
resource "upcloud_firewall_ruleset" "example" {
  name     = "example-ruleset"
  stateful = true
}

# Allow SSH access from a specific IP range
resource "upcloud_firewall_ruleset_rule" "allow_ssh" {
  ruleset_uuid = upcloud_firewall_ruleset.example.id
  action       = "accept"
  direction    = "in"
  family       = "IPv4"
  protocol     = "tcp"
  comment      = "Allow SSH from office network"

  source_address_start = "203.0.113.0"
  source_address_end   = "203.0.113.255"

  destination_port_start = 22
  destination_port_end   = 22

  position = 1
  enabled  = true
}

# Allow HTTPS traffic from anywhere
resource "upcloud_firewall_ruleset_rule" "allow_https" {
  ruleset_uuid = upcloud_firewall_ruleset.example.id
  action       = "accept"
  direction    = "in"
  family       = "IPv4"
  protocol     = "tcp"
  comment      = "Allow HTTPS traffic"

  destination_port_start = 443
  destination_port_end   = 443

  position = 2
  enabled  = true
}

# Allow ICMPv4 ping
resource "upcloud_firewall_ruleset_rule" "allow_icmp" {
  ruleset_uuid = upcloud_firewall_ruleset.example.id
  action       = "accept"
  direction    = "in"
  family       = "IPv4"
  protocol     = "icmp"
  comment      = "Allow ICMP ping"

  icmp_type = 8

  position = 3
  enabled  = true
}

# Block all other incoming traffic (explicit deny rule)
resource "upcloud_firewall_ruleset_rule" "deny_all" {
  ruleset_uuid = upcloud_firewall_ruleset.example.id
  action       = "reject"
  direction    = "in"
  family       = "IPv4"
  comment      = "Deny all other traffic"

  position = 100
  enabled  = true
}
