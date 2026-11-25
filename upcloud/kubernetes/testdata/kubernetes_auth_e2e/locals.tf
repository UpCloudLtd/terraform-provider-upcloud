locals {
  keycloak_url = "https://${one(kubernetes_service_v1.keycloak).status[0].load_balancer[0].ingress[0].hostname}"

  keycloak_token_request_map = {
    client_id     = "cluster"
    client_secret = one(keycloak_openid_client.client).client_secret
    grant_type    = "password"
    username      = one(keycloak_user.user).username
    password      = random_password.oidc_user_password.result
    scope         = "openid"
  }
  keycloak_token_request_list = [for k, v in local.keycloak_token_request_map : "${k}=${urlencode(v)}"]
  keycloak_token_request      = join("&", local.keycloak_token_request_list)

  id_token = jsondecode(one(data.http.token).response_body).id_token
}

output "id_token" {
  value     = local.id_token
  sensitive = true
}
