variable "basename" {
  default = "tf-acc-test-"
  type    = string
}

variable "network_cidr" {
  default = "172.23.45.0/24"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

variable "namespace" {
  type    = string
  default = "hello"
}

variable "enable_kubernetes_resources" {
  type    = bool
  default = false
}

variable "private_node_groups" {
  type    = bool
  default = false
}

locals {
  name_prefix = "${var.basename}k8s-e2e-${var.private_node_groups ? "priv-lb" : "publ-np"}-"
}

resource "upcloud_router" "main" {
  name = "${local.name_prefix}router"
}

resource "upcloud_network" "main" {
  name   = "${local.name_prefix}net"
  zone   = var.zone
  router = upcloud_router.main.id

  ip_network {
    address = var.network_cidr
    dhcp    = true
    dhcp_default_route = var.private_node_groups
    family  = "IPv4"
  }
}

resource "upcloud_gateway" "main" {
  # Only deploy NAT gateway when using private node-groups
  count = var.private_node_groups ? 1 : 0

  name     = "${local.name_prefix}gw"
  zone     = var.zone
  features = ["nat"]

  router {
    id = upcloud_router.main.id
  }
}

resource "upcloud_kubernetes_cluster" "main" {
  control_plane_ip_filter = ["0.0.0.0/0"]
  name                    = "${local.name_prefix}cluster"
  network                 = upcloud_network.main.id
  private_node_groups     = var.private_node_groups
  zone                    = var.zone
}

resource "upcloud_kubernetes_node_group" "default" {
  cluster       = upcloud_kubernetes_cluster.main.id
  node_count    = 1
  anti_affinity = true
  name          = "default"
  plan          = "1xCPU-2GB"
}

ephemeral "upcloud_kubernetes_cluster" "main" {
  id = upcloud_kubernetes_cluster.main.id
}

provider "kubernetes" {
  client_certificate     = ephemeral.upcloud_kubernetes_cluster.main.client_certificate
  client_key             = ephemeral.upcloud_kubernetes_cluster.main.client_key
  cluster_ca_certificate = ephemeral.upcloud_kubernetes_cluster.main.cluster_ca_certificate
  host                   = ephemeral.upcloud_kubernetes_cluster.main.host

  ignore_annotations = [
    "^service\\.beta\\.kubernetes\\.io\\/.*load.*balancer.*"
  ]
}

data "kubernetes_nodes" "this" {
  count = var.enable_kubernetes_resources ? 1 : 0

  depends_on = [upcloud_kubernetes_node_group.default]
}

resource "kubernetes_namespace_v1" "hello" {
  count = var.enable_kubernetes_resources ? 1 : 0

  metadata {
    name = var.namespace
  }
}

resource "kubernetes_deployment_v1" "hello" {
  count = var.enable_kubernetes_resources ? 1 : 0

  metadata {
    name      = "hello"
    namespace = var.namespace
    labels = {
      app = "hello"
    }
  }

  spec {
    selector {
      match_labels = {
        app = "hello"
      }
    }

    template {
      metadata {
        labels = {
          app = "hello"
        }
      }

      spec {
        container {
          image = "ghcr.io/upcloudltd/hello:latest"
          name  = "hello"
        }
      }
    }
  }
}

resource "kubernetes_service_v1" "hello" {
  count = var.enable_kubernetes_resources ? 1 : 0

  metadata {
    name      = "hello"
    namespace = var.namespace
  }
  spec {
    selector = {
      app = "hello"
    }

    port {
      port = var.private_node_groups ? 443 : 80
      target_port = 80
    }

    type = var.private_node_groups ? "LoadBalancer" : "NodePort"
  }
}

locals {
  addresses       = var.enable_kubernetes_resources ? data.kubernetes_nodes.this[0].nodes[0].status[0].addresses : []
  has_external_ip = var.enable_kubernetes_resources ? contains(local.addresses.*.type, "ExternalIP") : false
  external_ip     = local.has_external_ip ? local.addresses[index(local.addresses.*.type, "ExternalIP")].address : "localhost"
  port            = var.enable_kubernetes_resources ? kubernetes_service_v1.hello[0].spec[0].port[0].node_port : 8080
  lb_url = var.enable_kubernetes_resources && var.private_node_groups ? "https://${kubernetes_service_v1.hello[0].status[0].load_balancer[0].ingress[0].hostname}" : "localhost:8080"
  service_url     = var.private_node_groups ? local.lb_url : "http://${local.external_ip}:${local.port != null ? local.port : 8080}/"
}

data "http" "hello" {
  count = var.enable_kubernetes_resources ? 1 : 0

  url = local.service_url

  depends_on = [
    data.kubernetes_nodes.this,
    kubernetes_service_v1.hello,
  ]

  # Wait for the service to be ready:
  # - Max 5 minutes when using public node groups and NodePort service.
  # - Max 15 minutes when using private node groups and LoadBalancer service.
  retry {
    attempts     = var.private_node_groups ? 90 : 30
    min_delay_ms = 10e3
  }
}
