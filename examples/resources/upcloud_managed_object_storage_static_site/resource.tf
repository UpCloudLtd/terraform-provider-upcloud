resource "upcloud_managed_object_storage" "this" {
  name              = "example"
  region            = "europe-1"
  configured_status = "started"
}

resource "upcloud_managed_object_storage_static_site" "this" {
  service_uuid = "1201cd6f-20cd-44b6-939d-ae1c861769e7"

  bucket_name   = "example"
  bucket_prefix = "public/"

  error_pages = [{
    error_document = "404.html"
    status_code    = 404
  }]
}
