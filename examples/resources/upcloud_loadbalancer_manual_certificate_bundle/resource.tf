variable "certificate_path" {
  type = string
}

variable "private_key_path" {
  type = string
}

resource "upcloud_loadbalancer_manual_certificate_bundle" "lb_cb_m1" {
  name        = "lb-cb-m1-test"
  certificate = base64encode(file(var.certificate_path))
  private_key = base64encode(file(var.private_key_path))
}
