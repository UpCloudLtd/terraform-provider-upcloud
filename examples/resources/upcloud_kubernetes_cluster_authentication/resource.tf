# Create a network and a router for the Kubernetes cluster
resource "upcloud_router" "example" {
  name = "example-router"
}

resource "upcloud_network" "example" {
  name   = "example-network"
  zone   = "de-fra1"
  router = upcloud_router.example.id

  ip_network {
    address = "172.16.1.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

# Create a Kubernetes cluster
resource "upcloud_kubernetes_cluster" "example" {
  # Allow access to the cluster control plane from any external source.
  control_plane_ip_filter = ["0.0.0.0/0"]
  name                    = "exampleapp"
  network                 = upcloud_network.example.id
  zone                    = "de-fra1"
}

# Configure an OIDC issuer for the cluster
resource "upcloud_kubernetes_cluster_authentication" "example" {
  cluster    = upcloud_kubernetes_cluster.example.id
  name       = "example-issuer"
  issuer_url = "https://accounts.example.com"
  audiences  = ["example-client-id"]

  claim_mappings = {
    username = {
      claim  = "email"
      prefix = "oidc:"
    }
    groups = {
      claim  = "groups"
      prefix = "oidc:"
    }
  }

  claim_validation_rules = [
    {
      claim          = "hd"
      required_value = "example.com"
    }
  ]
}
