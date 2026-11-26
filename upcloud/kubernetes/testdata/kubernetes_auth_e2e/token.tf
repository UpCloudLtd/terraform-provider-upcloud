data "http" "token" {
  count = var.step >= 3 ? 1 : 0

  url    = "${local.keycloak_url}/realms/${one(keycloak_realm.realm).realm}/protocol/openid-connect/token"
  method = "POST"

  request_headers = {
    "Content-Type" = "application/x-www-form-urlencoded"
  }
  request_body = local.keycloak_token_request
}
