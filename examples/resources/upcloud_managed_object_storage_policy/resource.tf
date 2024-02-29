resource "upcloud_managed_object_storage" "this" {
  name              = "example"
  region            = "europe-1"
  configured_status = "started"
}

resource "upcloud_managed_object_storage_policy" "this" {
  name         = "example"
  description  = "example description"
  document     = "%7B%22Version%22%3A%20%222012-10-17%22%2C%20%20%22Statement%22%3A%20%5B%7B%22Action%22%3A%20%5B%22iam%3AGetUser%22%5D%2C%20%22Resource%22%3A%20%22%2A%22%2C%20%22Effect%22%3A%20%22Allow%22%2C%20%22Sid%22%3A%20%22editor%22%7D%5D%7D"
  service_uuid = upcloud_managed_object_storage.this.id
}
