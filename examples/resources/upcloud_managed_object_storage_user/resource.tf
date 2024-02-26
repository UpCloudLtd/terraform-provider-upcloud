resource "upcloud_managed_object_storage" "this" {
  name              = "example"
  region            = "europe-1"
  configured_status = "started"
}

resource "upcloud_managed_object_storage_user" "this" {
  username     = "example"
  service_uuid = upcloud_managed_object_storage.this.id
}
