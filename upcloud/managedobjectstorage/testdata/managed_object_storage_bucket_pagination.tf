variable "prefix" {
  type = string
}

variable "region" {
  default = "europe-1"
  type    = string
}

locals {
  bucket_names = [for index in range(12) : format("%s-%02d", var.prefix, index)]
}

resource "upcloud_managed_object_storage" "this" {
  name              = var.prefix
  region            = var.region
  configured_status = "started"
}

resource "upcloud_managed_object_storage_bucket" "this" {
  count = length(local.bucket_names)

  service_uuid = upcloud_managed_object_storage.this.id
  name         = local.bucket_names[count.index]
}