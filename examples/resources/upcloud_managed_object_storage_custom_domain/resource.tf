resource "upcloud_managed_object_storage" "this" {
  name              = "object-storage-custom-domain-example"
  region            = "europe-1"
  configured_status = "started"

  network {
    family = "IPv4"
    name   = "public"
    type   = "public"
  }
}

data "upcloud_load_balancer_dns_challenge_domain" "this" {}

// Before creating the custom domain, configure the DNS settings for your custom domain. For example, if your custom domain is objects.example.com, you should configure the following DNS records:
// - "_acme-challenge.objects IN CNAME ${data.upcloud_load_balancer_dns_challenge_domain.this.domain}"
// - "objects IN CNAME ${[for i in upcloud_managed_object_storage.this.endpoint: i.domain_name if i.type == "public"][0]}"
// - "*.objects IN CNAME ${[for i in upcloud_managed_object_storage.this.endpoint: i.domain_name if i.type == "public"][0]}"
resource "upcloud_managed_object_storage_custom_domain" "this" {
  service_uuid = upcloud_managed_object_storage.this.id
  domain_name  = "objects.example.com"
}
