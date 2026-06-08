# Create a stateful firewall ruleset with default DNS rules enabled
resource "upcloud_firewall_ruleset" "example" {
  name        = "example-ruleset"
  description = "Example firewall ruleset for production servers"
  stateful    = true

  # Enable default DNS rules to allow DNS traffic
  default_dns_rules_enabled = true

  # Add labels for organization
  labels = {
    environment = "production"
    managed_by  = "terraform"
  }
}

# Create a firewall ruleset attached to a specific server
resource "upcloud_server" "example" {
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"

  template {
    storage = "Ubuntu Server 24.04 LTS (Noble Numbat)"
  }

  network_interface {
    type = "utility"
  }
}

resource "upcloud_firewall_ruleset" "server_ruleset" {
  name        = "server-ruleset"
  description = "Firewall rules for example server"
  stateful    = true
  server_uuid = upcloud_server.example.id
}
