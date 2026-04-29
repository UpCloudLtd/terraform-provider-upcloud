variable "prefix" {
  default = "tf-acc-test-objstov2-"
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
  name              = "tf-acc-test-objstov2-data-source"
  region            = var.region
  configured_status = "started"
}

data "upcloud_managed_object_storage_policies" "this" {
  service_uuid = upcloud_managed_object_storage.this.id
}
