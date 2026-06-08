variable "ruleset_name" {
  type = string
}

resource "upcloud_firewall_ruleset" "test" {
  name                      = var.ruleset_name
  description               = "Test firewall ruleset"
  enabled                   = true
  stateful                  = true
  default_dns_rules_enabled = false
}
