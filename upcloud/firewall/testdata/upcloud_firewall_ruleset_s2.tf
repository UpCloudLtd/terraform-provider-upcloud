variable "ruleset_name" {
  type = string
}

resource "upcloud_firewall_ruleset" "test" {
  name                      = var.ruleset_name
  description               = "Updated description"
  enabled                   = false
  stateful                  = true
  default_dns_rules_enabled = true
}
