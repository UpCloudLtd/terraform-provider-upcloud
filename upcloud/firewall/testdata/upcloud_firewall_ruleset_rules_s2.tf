variable "ruleset_name" {
  type = string
}

resource "upcloud_firewall_ruleset" "test" {
  name    = var.ruleset_name
  enabled = true

  rules = [
    {
      action                 = "accept"
      direction              = "in"
      family                 = "IPv4"
      protocol               = "tcp"
      destination_port_start = 443
      destination_port_end   = 443
      enabled                = true
      comment                = "Allow HTTPS"
    },
    {
      action                 = "accept"
      direction              = "in"
      family                 = "IPv4"
      protocol               = "tcp"
      destination_port_start = 80
      destination_port_end   = 80
      enabled                = true
      comment                = "Allow HTTP"
    },
    {
      action                 = "accept"
      direction              = "in"
      family                 = "IPv4"
      protocol               = "tcp"
      destination_port_start = 22
      destination_port_end   = 22
      enabled                = true
      comment                = "Allow SSH"
    },
  ]
}
