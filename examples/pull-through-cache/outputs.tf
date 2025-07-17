output "token" {
  sensitive = true
  value     = random_password.token.result
}

output "url" {
  value = upcloud_loadbalancer.this.networks[0].dns_name
}
