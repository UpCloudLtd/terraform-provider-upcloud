resource "upcloud_kubernetes_cluster_authentication" "keycloak" {
  count = var.step >= 4 ? 1 : 0

  cluster    = upcloud_kubernetes_cluster.main.id
  name       = "keycloak"
  issuer_url = "${local.keycloak_url}/realms/${one(keycloak_realm.realm).realm}"
  audiences  = ["cluster"]

  claim_mappings = {
    username = {
      claim  = "preferred_username"
      prefix = "oidc:"
    }
  }
}

ephemeral "upcloud_kubernetes_kubeconfig" "oidc" {
  count = var.step >= 4 ? 1 : 0

  cluster        = upcloud_kubernetes_cluster.main.id
  authentication = one(upcloud_kubernetes_cluster_authentication.keycloak).name
  audience       = "cluster"
}
