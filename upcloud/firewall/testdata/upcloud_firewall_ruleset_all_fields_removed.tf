variable "ruleset_name" {
  type = string
}

# The same rules as upcloud_firewall_ruleset_all_fields.tf, with every optional field left out.
resource "upcloud_firewall_ruleset" "test" {
  name = var.ruleset_name

  rules = [
    {
      action    = "accept"
      direction = "in"
      family    = "IPv4"
    },
    {
      action    = "accept"
      direction = "in"
      family    = "IPv4"
    },
    {
      action    = "accept"
      direction = "in"
      family    = "IPv4"
    },
  ]
}
