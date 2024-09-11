data "upcloud_load_balancer_dns_challenge_domain" "this" {}

// Configure _acme-challenge by creating "_acme-challenge.objects IN CNAME ${data.upcloud_load_balancer_dns_challenge_domain.this.domain}" DNS record.
