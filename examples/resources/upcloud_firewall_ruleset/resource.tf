# Create a stateful firewall ruleset with default DNS rules enabled
resource "upcloud_firewall_ruleset" "example" {
  name        = "example-ruleset"
  description = "Example firewall ruleset for production servers"

  # Enable default DNS rules to allow DNS traffic
  default_dns_rules_enabled = true

  # Add labels for organization
  labels = {
    environment = "production"
    managed_by  = "terraform"
  }
}
