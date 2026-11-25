ephemeral "upcloud_kubernetes_kubeconfig" "oidc" {
  count = var.step >= 5 ? 1 : 0

  cluster        = upcloud_kubernetes_cluster.main.id
  authentication = one(upcloud_kubernetes_cluster_authentication.keycloak).name
  audience       = "cluster"

  depends_on = [
    time_sleep.wait_for_auth_config,
  ]
}

resource "kubernetes_namespace_v1" "hello" {
  provider = kubernetes.oidc
  count    = var.step >= 5 ? 1 : 0

  metadata {
    name = "hello"
  }
}

resource "kubernetes_deployment_v1" "hello" {
  provider = kubernetes.oidc
  count    = var.step >= 5 ? 1 : 0

  metadata {
    name      = "hello"
    namespace = one(kubernetes_namespace_v1.hello).metadata[0].name
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
