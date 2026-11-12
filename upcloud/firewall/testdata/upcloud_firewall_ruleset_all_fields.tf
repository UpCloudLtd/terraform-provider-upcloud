variable "ruleset_name" {
  type = string
}

variable "enabled" {
  type    = bool
  default = true
}

# Sets every optional field. Rule enabled is left out, because the API might ignore enabled = false on rules.
resource "upcloud_firewall_ruleset" "test" {
  name                      = var.ruleset_name
  description               = "test"
  enabled                   = var.enabled
  default_dns_rules_enabled = true

  rules = [
    {
      action                   = "accept"
      direction                = "in"
      family                   = "IPv4"
      protocol                 = "icmp"
      icmp_type                = 13
      comment                  = "allow icmp"
      source_address_cidr      = "10.0.0.0/8"
      destination_address_cidr = "10.0.0.0/8"
    },
    {
      action                    = "accept"
      direction                 = "in"
      family                    = "IPv4"
      protocol                  = "tcp"
      source_port_start         = 1024
      source_port_end           = 2048
      destination_port_start    = 1024
      destination_port_end      = 2048
      source_address_start      = "10.0.0.1"
      source_address_end        = "10.0.0.2"
      destination_address_start = "10.0.0.1"
      destination_address_end   = "10.0.0.2"
    },
    {
      action    = "accept"
      direction = "in"
      family    = "IPv4"
      protocol  = "all"
    },
  ]
}
