variable "ruleset_name" {
  type = string
}

resource "upcloud_firewall_ruleset" "test" {
  name     = var.ruleset_name
  stateful = true
}

resource "upcloud_firewall_ruleset_rule" "test" {
  ruleset_uuid           = upcloud_firewall_ruleset.test.id
  action                 = "accept"
  direction              = "in"
  family                 = "IPv4"
  protocol               = "tcp"
  source_port_start      = 1024
  source_port_end        = 65535
  destination_port_start = 8000
  destination_port_end   = 9000
}
