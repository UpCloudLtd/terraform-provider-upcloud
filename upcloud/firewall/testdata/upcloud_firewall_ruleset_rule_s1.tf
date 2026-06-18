variable "ruleset_name" {
  type = string
}

resource "upcloud_firewall_ruleset" "test" {
  name     = var.ruleset_name
  stateful = true
}

resource "upcloud_firewall_ruleset_rule" "allow_http" {
  ruleset_uuid           = upcloud_firewall_ruleset.test.id
  action                 = "accept"
  direction              = "in"
  family                 = "IPv4"
  protocol               = "tcp"
  destination_port_start = 80
  destination_port_end   = 80
  enabled                = true
  comment                = "Allow HTTP traffic"
}
