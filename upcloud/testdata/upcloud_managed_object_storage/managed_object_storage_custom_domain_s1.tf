variable "prefix" {
  default = "tf-acc-test-objstov2-custom-domain-"
  type    = string
}

variable "zone" {
  default = "se-sto1"
  type    = string
}

variable "region" {
  default = "europe-3"
  type    = string
}

resource "upcloud_managed_object_storage" "this" {
  name              = "${var.prefix}objsto"
  region            = var.region
  configured_status = "started"

  network {
    family = "IPv4"
    name   = "public"
    type   = "public"
  }
}

resource "upcloud_managed_object_storage_custom_domain" "this" {
  service_uuid = upcloud_managed_object_storage.this.id
  domain_name  = "objects.example.com"
}
