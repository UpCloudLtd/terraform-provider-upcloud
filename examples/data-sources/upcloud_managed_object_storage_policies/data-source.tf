resource "upcloud_managed_object_storage" "this" {
  name              = "example"
  region            = "europe-1"
  configured_status = "started"
}

data "upcloud_Managed_object_storage_policies" "this" {
  service_uuid = upcloud_managed_object_storage.this.id
}
