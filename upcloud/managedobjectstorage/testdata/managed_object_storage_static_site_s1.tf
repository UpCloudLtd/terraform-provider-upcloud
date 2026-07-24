variable "prefix" {
  default = "tf-acc-test-objstov2-static-site-"
  type    = string
}

variable "region" {
  default = "europe-3"
  type    = string
}

resource "upcloud_managed_object_storage" "this" {
  name              = "${var.prefix}static"
  region            = var.region
  configured_status = "started"

  network {
    family = "IPv4"
    name   = "public"
    type   = "public"
  }
}

resource "upcloud_managed_object_storage_bucket" "this" {
  service_uuid = upcloud_managed_object_storage.this.id
  name         = "website"
}

resource "upcloud_managed_object_storage_static_site" "this" {
  service_uuid   = upcloud_managed_object_storage.this.id
  bucket_name    = upcloud_managed_object_storage_bucket.this.name
  bucket_prefix  = ""
  index_document = "index.html"
  spa_mode       = false
  enabled        = true
}

