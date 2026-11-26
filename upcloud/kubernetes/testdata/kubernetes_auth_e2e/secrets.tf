resource "random_password" "keycloak_admin_client_secret" {
  length  = 64
  special = false
}

resource "random_password" "keycloak_admin_password" {
  length  = 16
  special = false
}

resource "random_password" "oidc_user_password" {
  length  = 16
  special = false
}
