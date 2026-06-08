variable "ruleset_name" {
  type = string
}

resource "upcloud_firewall_ruleset" "test" {
  name     = var.ruleset_name
  stateful = true
}

resource "upcloud_firewall_ruleset_rule" "test" {
  ruleset_uuid              = upcloud_firewall_ruleset.test.id
  action                    = "accept"
  direction                 = "in"
  family                    = "IPv4"
  protocol                  = "tcp"
  source_address_start      = "192.168.1.10"
  source_address_end        = "192.168.1.20"
  destination_address_start = "10.0.0.1"
  destination_address_end   = "10.0.0.254"
}
