# Create a network for the Kubernetes cluster
resource "upcloud_network" "example" {
  name = "example-network"
  zone = "de-fra1"
  ip_network {
    address = "172.16.1.0/24"
    dhcp    = true
    family  = "IPv4"
  }

  # UpCloud Kubernetes Service will add a router to this network to ensure cluster networking is working as intended.
  # You need to ignore changes to it, otherwise TF will attempt to detach the router on subsequent applies
  lifecycle {
    ignore_changes = [router]
  }
}

# Create a Kubernetes cluster
resource "upcloud_kubernetes_cluster" "example" {
  name    = "exampleapp"
  network = upcloud_network.example.id
  zone    = "de-fra1"
}

# Kubernetes cluster with private node groups requires network that is routed through NAT gateway.
resource "upcloud_router" "example2" {
  name = "example2-router"
}

resource "upcloud_gateway" "example2" {
  name     = "example2-nat-gateway"
  zone     = "de-fra1"
  features = ["nat"]

  router {
    id = upcloud_router.example2.id
  }
}

resource "upcloud_network" "example2" {
  name = "example2-network"
  zone = "de-fra1"
  ip_network {
    address            = "10.10.0.0/24"
    dhcp               = true
    family             = "IPv4"
    dhcp_default_route = true
  }
  router = upcloud_router.example2.id
}

resource "upcloud_kubernetes_cluster" "example2" {
  name                = "example2-cluster"
  network             = upcloud_network.example2.id
  zone                = "de-fra1"
  plan                = "production-small"
  private_node_groups = true
}

