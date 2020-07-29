---
layout: "upcloud"
page_title: "Provider: UpCloud"
sidebar_current: "docs-upcloud-index"
description: |-
  UpCloud
---

# UpCloud Provider

The UpCloud Terraform Provider enables organisations to control resources hosted on the UpCloud hosting platform. 

## Example Usage

```hcl
terraform {
  required_version = ">= 0.12.0"
}

provider "upcloud" {
  username = "<YOUR username>"
  password = "<Your password>"
}

resource "upcloud_server" "myserver" {
  # ...
}
```

## Argument Reference

The following arguments are supported:

* `username` - (Optional) The UpCloud username with API access. It must be provided, but
  it can also be sourced from the `UPCLOUD_USERNAME` environment variable.

* `password` - (Optional) The Password for UpCloud API user. It must be provided, but
  it can also be sourced from the `UPCLOUD_PASSWORD` environment variable.