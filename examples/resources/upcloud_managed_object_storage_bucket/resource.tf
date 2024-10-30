resource "upcloud_managed_object_storage" "example" {
  name              = "bucket-example-objstov2"
  region            = "europe-1"
  configured_status = "started"
}

resource "upcloud_managed_object_storage_bucket" "example" {
  service_uuid = upcloud_managed_object_storage.example.id
  name         = "bucket"
}
