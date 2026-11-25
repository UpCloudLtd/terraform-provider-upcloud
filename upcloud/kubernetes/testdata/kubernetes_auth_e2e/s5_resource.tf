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
    namespace = "hello"
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
