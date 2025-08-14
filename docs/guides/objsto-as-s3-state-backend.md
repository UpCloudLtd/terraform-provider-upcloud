---
page_title: Managed Object Storage as S3 State Backend
---

# Managed Object Storage as S3 State Backend

~> The S3 backends support for non AWS S3 implementations is not tested by the OpenTofu or Terraform teams, so there might be issues when OpenTofu and Terraform implementations adapt to new functionality in AWS S3. Another option is to use UpCloud Managed Databases for PostgreSQL database as the backend.

This guide presents an example of how to configure UpCloud Managed Object Storage as a state backend for OpenTofu or Terraform using the S3 backend.

```terraform
terraform {
  # Other configuration, such as required_providers, omitted.

  backend "s3" {
    # Define the name of your bucket and the key for the state file.
    bucket = "example-bucket"
    key    = "example.tfstate"


    # We need to skip AWS specific checks and use path style URLs for
    # UpCloud Managed Object Storage.
    skip_requesting_account_id = true
    skip_credentials_validation = true
    skip_metadata_api_check = true
    skip_region_validation = true
    skip_s3_checksum = true
    use_path_style = true

    # Set the region to match your Managed Object Storage instance.
    region = "europe-1"

    # Configure the endpoints for your Managed Object Storage instance.
    endpoints = {
        s3 = "https://example.upcloudobjects.com"
        iam = "https://example.upcloudobjects.com:4443/iam"
        sts = "https://example.upcloudobjects.com:4443/sts"
    }

    # Credentials can be defined either in configuration or with
    # AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment
    # variables.
  }
}
```

Note that with some versions of OpenTofu or Terraform the S3 backend might do additional integrity checks, even when `skip_s3_checksum` option is set to `true`, because of changes to default behavior of AWS Go SDK. This is visible as `XAmzContent*Mismatch` errors when saving the state. To disable these checks, set `request_checksum_calculation` and `response_checksum_validation` options to `when_required`. This can be done, for example, with environment variables:

```sh
export AWS_REQUEST_CHECKSUM_CALCULATION=when_required
export AWS_RESPONSE_CHECKSUM_VALIDATION=when_required
```
