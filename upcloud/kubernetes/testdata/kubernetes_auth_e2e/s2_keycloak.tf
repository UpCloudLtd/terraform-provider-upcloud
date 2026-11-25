resource "kubernetes_namespace_v1" "keycloak" {
  provider = kubernetes.cluster
  count    = var.step >= 2 ? 1 : 0

  metadata {
    name = "keycloak"
  }
}

resource "kubernetes_deployment_v1" "keycloak" {
  provider = kubernetes.cluster
  count    = var.step >= 2 ? 1 : 0

  metadata {
    name      = "keycloak"
    namespace = one(kubernetes_namespace_v1.keycloak).metadata[0].name
    labels = {
      app = "keycloak"
    }
  }

  spec {
    selector {
      match_labels = {
        app = "keycloak"
      }
    }

    template {
      metadata {
        labels = {
          app = "keycloak"
        }
      }

      spec {
        container {
          image = "quay.io/keycloak/keycloak:26.8.2"
          name  = "keycloak"
          args  = ["start-dev"]

          env {
            name  = "KC_HTTP_HOST"
            value = "0.0.0.0"
          }

          env {
            name  = "KC_PROXY_HEADERS"
            value = "xforwarded"
          }

          env {
            name  = "KC_BOOTSTRAP_ADMIN_USERNAME"
            value = "admin"
          }

          env {
            name  = "KC_BOOTSTRAP_ADMIN_PASSWORD"
            value = random_password.keycloak_admin_password.result
          }

          env {
            name  = "KC_BOOTSTRAP_ADMIN_CLIENT_ID"
            value = "admin"
          }

          env {
            name  = "KC_BOOTSTRAP_ADMIN_CLIENT_SECRET"
            value = random_password.keycloak_admin_client_secret.result
          }
        }
      }
    }
  }
}

resource "kubernetes_service_v1" "keycloak" {
  provider = kubernetes.cluster
  count    = var.step >= 2 ? 1 : 0

  metadata {
    name      = "keycloak"
    namespace = one(kubernetes_namespace_v1.keycloak).metadata[0].name
    annotations = {
      "service.beta.kubernetes.io/upcloud-load-balancer-config" = jsonencode({
        name = "${local.name_prefix}lb"
        frontends = [
          {
            rules = [{
              name     = "set-forwarded-headers"
              priority = 90
              actions = [
                {
                  type                         = "set_forwarded_headers"
                  action_set_forwarded_headers = {}
                }
              ]
            }]
          }
        ]
      })
    }
  }

  spec {
    selector = {
      app = "keycloak"
    }

    port {
      port        = 443
      target_port = 8080
    }

    type = "LoadBalancer"
  }
}

data "http" "keycloak" {
  count = var.step >= 2 ? 1 : 0

  url = "${local.keycloak_url}/realms/master"

  depends_on = [
    kubernetes_deployment_v1.keycloak,
    kubernetes_service_v1.keycloak,
  ]

  # Wait for the load balancer and Keycloak to be ready before proceeding to the next step. Max 15 minutes.
  retry {
    attempts     = 90
    min_delay_ms = 10e3
  }

  lifecycle {
    postcondition {
      condition     = self.status_code == 200
      error_message = "Keycloak is not ready, status code: ${self.status_code}."
    }
  }
}
