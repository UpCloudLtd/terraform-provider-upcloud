variable "ruleset_name" {
  type = string
}

resource "upcloud_firewall_ruleset" "test" {
  name     = var.ruleset_name

  labels = {
    env        = "test"
    managed-by = "terraform"
    purpose    = "acceptance-test"
  }
}
