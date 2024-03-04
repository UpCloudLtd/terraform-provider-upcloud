variable "prefix" {
  default = "tf-acc-test-objstov2-"
  type    = string
}

variable "region" {
  default = "europe-1"
  type    = string
}

resource "upcloud_managed_object_storage" "user" {
  name              = "${var.prefix}user"
  region            = var.region
  configured_status = "started"
}

resource "upcloud_managed_object_storage_policy" "user" {
  description  = "${var.prefix}user-desc"
  name         = "${var.prefix}user"
  document     = "%7B%22Version%22%3A%222012-10-17%22%2C%22Statement%22%3A%5B%7B%22Action%22%3A%5B%22iam%3AGetUser%22%5D%2C%22Resource%22%3A%22*%22%2C%22Effect%22%3A%22Allow%22%2C%22Sid%22%3A%22editor%22%7D%5D%7D"
  service_uuid = upcloud_managed_object_storage.user.id
}

resource "upcloud_managed_object_storage_user" "user" {
  username     = "${var.prefix}user"
  service_uuid = upcloud_managed_object_storage.user.id
}

resource "upcloud_managed_object_storage_user_access_key" "user" {
  username     = upcloud_managed_object_storage_user.user.username
  service_uuid = upcloud_managed_object_storage.user.id
  status       = "Active"
}

resource "upcloud_managed_object_storage_user_policy" "user" {
  name         = upcloud_managed_object_storage_policy.user.name
  username     = upcloud_managed_object_storage_user.user.username
  service_uuid = upcloud_managed_object_storage.user.id
}