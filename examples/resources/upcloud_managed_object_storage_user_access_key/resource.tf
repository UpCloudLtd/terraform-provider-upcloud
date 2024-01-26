resource "upcloud_managed_object_storage" "this" {
  name              = "example"
  region            = "europe-1"
  configured_status = "started"
  users             = ["example"]
}

resource "upcloud_managed_object_storage_user_access_key" "this" {
  name         = "accesskey"
  enabled      = true
  username     = "example"
  service_uuid = upcloud_managed_object_storage.this.id
}
