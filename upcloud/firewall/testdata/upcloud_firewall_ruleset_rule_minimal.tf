variable "ruleset_name" {
  type = string
}

resource "upcloud_firewall_ruleset" "test" {
  name     = var.ruleset_name
  stateful = true
}

resource "upcloud_firewall_ruleset_rule" "test" {
  ruleset_uuid = upcloud_firewall_ruleset.test.id
  action       = "accept"
  direction    = "in"
  family       = "IPv4"
}
