terraform {
  required_providers {
    http = {
      source  = "hashicorp/http"
      version = "~> 3.6"
    }
    keycloak = {
      source  = "keycloak/keycloak"
      version = "~> 5.11"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 3.5"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.9"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.14"
    }
  }
}

provider "kubernetes" {
  alias = "cluster"

  client_certificate     = ephemeral.upcloud_kubernetes_cluster.main.client_certificate
  client_key             = ephemeral.upcloud_kubernetes_cluster.main.client_key
  cluster_ca_certificate = ephemeral.upcloud_kubernetes_cluster.main.cluster_ca_certificate
  host                   = ephemeral.upcloud_kubernetes_cluster.main.host

  ignore_annotations = [
    "^service\\.beta\\.kubernetes\\.io\\/.*load.*balancer.*"
  ]
}

provider "keycloak" {
  client_id     = "admin"
  client_secret = random_password.keycloak_admin_client_secret.result
  url           = local.keycloak_url
  initial_login = false
}

provider "kubernetes" {
  alias = "oidc"

  cluster_ca_certificate = ephemeral.upcloud_kubernetes_cluster.main.cluster_ca_certificate
  host                   = ephemeral.upcloud_kubernetes_cluster.main.host
  token                  = local.id_token

  ignore_annotations = [
    "^service\\.beta\\.kubernetes\\.io\\/.*load.*balancer.*"
  ]
}
