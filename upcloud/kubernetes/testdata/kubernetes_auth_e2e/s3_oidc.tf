resource "keycloak_realm" "realm" {
  count = var.step >= 3 ? 1 : 0

  realm        = "cloud"
  enabled      = true
  ssl_required = "external"

  # Keycloak needs to be running when the realm is created or destroyed. Without this, Terraform could destroy the deployment before the realm.
  depends_on = [
    kubernetes_deployment_v1.keycloak,
    kubernetes_service_v1.keycloak,
  ]
}

resource "keycloak_openid_client" "client" {
  count = var.step >= 3 ? 1 : 0

  realm_id    = one(keycloak_realm.realm).id
  client_id   = "cluster"
  enabled     = true
  access_type = "CONFIDENTIAL"

  direct_access_grants_enabled = true
}

resource "keycloak_user" "user" {
  count = var.step >= 3 ? 1 : 0

  realm_id = one(keycloak_realm.realm).id
  enabled  = true

  username       = var.oidc_username
  email          = "${var.oidc_username}@example.com"
  email_verified = true
  first_name     = "Teppo"
  last_name      = "Tulppu"

  initial_password {
    value     = random_password.oidc_user_password.result
    temporary = false
  }

  # Do not delete the role binding before the user.
  depends_on = [
    kubernetes_cluster_role_binding_v1.oidc
  ]
}

resource "kubernetes_cluster_role_binding_v1" "oidc" {
  provider = kubernetes.cluster
  count    = var.step >= 3 ? 1 : 0

  metadata {
    name = "oidc-role-binding"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "cluster-admin"
  }
  subject {
    api_group = "rbac.authorization.k8s.io"
    kind      = "User"
    name      = "oidc:${var.oidc_username}"
  }
}
