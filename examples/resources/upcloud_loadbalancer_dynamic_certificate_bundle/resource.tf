resource "upcloud_loadbalancer_dynamic_certificate_bundle" "lb_cb_d1" {
  name = "lb-cb-d1-test"
  hostnames = [
    "example.com",
    "app.example.net",
  ]
  key_type = "rsa"
}
