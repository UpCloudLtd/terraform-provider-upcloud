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
  comment                = "Allow HTTP"
  position               = 1
}

resource "upcloud_firewall_ruleset_rule" "allow_https" {
  ruleset_uuid           = upcloud_firewall_ruleset.test.id
  action                 = "accept"
  direction              = "in"
  family                 = "IPv4"
  protocol               = "tcp"
  destination_port_start = 443
  destination_port_end   = 443
  enabled                = true
  comment                = "Allow HTTPS"
  position               = 2
}

resource "upcloud_firewall_ruleset_rule" "allow_ssh" {
  ruleset_uuid           = upcloud_firewall_ruleset.test.id
  action                 = "accept"
  direction              = "in"
  family                 = "IPv4"
  protocol               = "tcp"
  destination_port_start = 22
  destination_port_end   = 22
  enabled                = true
  comment                = "Allow SSH"
  position               = 3
}

resource "upcloud_firewall_ruleset_rule" "drop_all" {
  ruleset_uuid = upcloud_firewall_ruleset.test.id
  action       = "drop"
  direction    = "in"
  family       = "IPv4"
  enabled      = true
  comment      = "Drop all other traffic"
  position     = 100
}
