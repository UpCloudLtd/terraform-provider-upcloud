variable "prefix" {
  default = "tf-acc-test-objsto2-"
  type    = string
}

variable "zone" {
  default = "fi-hel1"
  type    = string
}

variable "region" {
  default = "europe-2"
  type    = string
}

resource "upcloud_managed_object_storage" "this" {
  name              = "tf-acc-test-objsto2-data-source"
  region            = var.region
  configured_status = "started"
}

data "upcloud_managed_object_storage_policies" "this" {
  service_uuid = upcloud_managed_object_storage.this.id
}
