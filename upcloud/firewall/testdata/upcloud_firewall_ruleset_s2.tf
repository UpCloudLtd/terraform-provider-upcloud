variable "ruleset_name" {
  type = string
}

resource "upcloud_firewall_ruleset" "test" {
  name                      = var.ruleset_name
  description               = "Updated description"
  enabled                   = false
  default_dns_rules_enabled = true
}
