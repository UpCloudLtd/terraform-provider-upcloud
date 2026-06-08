variable "region" {
  default = "europe-3"
}

resource "upcloud_managed_object_storage" "this" {
  name              = "tf-acc-test-objsto-policy-version"
  region            = var.region
  configured_status = "started"
}

resource "upcloud_managed_object_storage_policy" "this" {
  name = "versioned-policy"

  document = urlencode(jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid      = "ReadOnly"
        Effect   = "Allow"
        Action   = ["s3:GetObject"]
        Resource = "arn:aws:s3:::bucket/*"
      }
    ]
  }))

  service_uuid = upcloud_managed_object_storage.this.id
}
