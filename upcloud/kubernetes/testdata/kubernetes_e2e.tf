variable "basename" {
  default = "tf-acc-test-"
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

locals {
  name_prefix = "${var.basename}k8s-e2e-"
}

resource "upcloud_router" "main" {
  name = "${local.name_prefix}router"
}

resource "upcloud_network" "main" {
  name   = "${local.name_prefix}net"
  zone   = var.zone
  router = upcloud_router.main.id

  ip_network {
    address = "172.23.45.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_kubernetes_cluster" "main" {
  control_plane_ip_filter = ["0.0.0.0/0"]
  name                    = "${local.name_prefix}cluster"
  network                 = upcloud_network.main.id
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
      port        = 80
      target_port = 80
    }

    type = "NodePort"
  }
}

locals {
  addresses       = var.enable_kubernetes_resources ? data.kubernetes_nodes.this[0].nodes[0].status[0].addresses : []
  has_external_ip = var.enable_kubernetes_resources ? contains(local.addresses.*.type, "ExternalIP") : false
  external_ip     = local.has_external_ip ? local.addresses[index(local.addresses.*.type, "ExternalIP")].address : "localhost"
  port            = var.enable_kubernetes_resources ? kubernetes_service_v1.hello[0].spec[0].port[0].node_port : 8080
  service_url     = "http://${local.external_ip}:${local.port != null ? local.port : 8080}/"
}

data "http" "hello" {
  count = var.enable_kubernetes_resources ? 1 : 0

  url = local.service_url

  depends_on = [
    data.kubernetes_nodes.this,
    kubernetes_service_v1.hello,
  ]

  # Wait 5 minutes for the service to be ready.
  retry {
    attempts     = 30
    min_delay_ms = 10e3
  }
}
