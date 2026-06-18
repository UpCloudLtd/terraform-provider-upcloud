variable "ruleset_name" {
  type = string
}

resource "upcloud_firewall_ruleset" "test" {
  name     = var.ruleset_name
  stateful = true
}

resource "upcloud_firewall_ruleset_rule" "test" {
  ruleset_uuid             = upcloud_firewall_ruleset.test.id
  action                   = "accept"
  direction                = "in"
  family                   = "IPv4"
  protocol                 = "tcp"
  source_address_cidr      = "192.168.1.0/24"
  destination_address_cidr = "10.0.0.0/8"
}
