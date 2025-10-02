variable "prefix" {
  default = "tf-acc-test-objstov2-iam-"
  type    = string
}

variable "region" {
  default = "europe-1"
  type    = string
}

resource "upcloud_managed_object_storage" "user" {
  name              = "${var.prefix}objsto"
  region            = var.region
  configured_status = "started"
}

resource "upcloud_managed_object_storage_policy" "user" {
  description = "Allow get access to the users."
  name        = "get-user-policy"
  document = urlencode(jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["iam:GetUser"]
        Resource = "*"
        Effect   = "Allow"
        Sid      = "ReadUser"
      }
    ]
  }))
  service_uuid = upcloud_managed_object_storage.user.id
}

resource "upcloud_managed_object_storage_policy" "escape" {
  name = "put-object-policy"
  document = urlencode(jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["s3:PutObject"]
        Resource = "arn:aws:s3:::bucket/*"
        Effect   = "Allow"
        Sid      = "WriteObject"
      }
    ]
  }))
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