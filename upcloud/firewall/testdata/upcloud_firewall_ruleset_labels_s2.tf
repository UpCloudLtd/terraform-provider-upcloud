variable "ruleset_name" {
  type = string
}

resource "upcloud_firewall_ruleset" "test" {
  name     = var.ruleset_name
  stateful = true

  labels = {
    env        = "production"
    managed-by = "terraform"
  }
}
