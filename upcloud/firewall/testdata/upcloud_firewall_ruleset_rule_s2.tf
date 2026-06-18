variable "ruleset_name" {
  type = string
}

resource "upcloud_firewall_ruleset" "test" {
  name     = var.ruleset_name
  stateful = true
}

resource "upcloud_firewall_ruleset_rule" "allow_http" {
  ruleset_uuid           = upcloud_firewall_ruleset.test.id
  action                 = "drop"
  direction              = "in"
  family                 = "IPv4"
  protocol               = "tcp"
  destination_port_start = 8080
  destination_port_end   = 8080
  enabled                = false
  comment                = "Drop alternative HTTP port"
}
