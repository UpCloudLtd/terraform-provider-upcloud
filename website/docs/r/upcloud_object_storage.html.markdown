---
layout: "upcloud"
page_title: "UpCloud: resource_upcloud_object_storage"
sidebar_current: "docs-upcloud-resource-upcloud_object_storage"
description: |-
Provides a UpCloud Object Storage management
---

# upcloud_object_storage

This resource represents an UpCloud Object Storage instance, which provides S3 compatible storage.

## Example Usage

The following example represents an object storage instance called `storage-name` in the `fi-hel2` zone, with 2 buckets called `"products"` and `"images"`.

```hcl
resource "upcloud_object_storage" "my_object_storage" {
  size  = 250               # allocate up to 250GB of storage
  name = "storage-name"     # the instance name, will form part of the url used to access the storage
                            # instance so must conform to host naming rules.
  zone  = "fi-hel2"         # The zone in wgich to create the instance
  access_key = "admin"      # The access key/username used to access the storage instance
  secret_key = "changeme"   # The secret key/password used to access the storage instance
  description = "catalogue"

  # Create a bucket called "products"
  bucket {
    name = "products"
  }

  # Create a bucket called "images"
  bucket {
    name = "images"
  }
}
```

## Argument Reference

The following arguments are supported:

* `size` - (Required) The maximum amount of storage to allocate to the object storage instance in gigabytes. Valid values are `250`, `500` and `1000`

* `name` - (Required) The name of the Object Storage instance, this value forms part of the url used to access the storage instance so must conform to host naming rules.

* `zone` - (Required) - The zone in which the server will be hosted, e.g. fi-hel2. See [Zones API](https://developers.upcloud.com/1.3/5-zones/)

* `access_key` - (Required) - The access key/username used to access the storage instance. Must be a string between 4 and 255 characters in length.

* `secret_key` - (Required) - The secret key/password used to access the storage instance.  Must be a string between 8 and 255 characters in length.

* `description` - (Optional) - A user defined string containing test to associate with the storage instance.

### Bucket

Each of `bucket` blocks represents a single bucket in the object storage instance and supports the following arguments:

* `name` - (Required) The name of the bucket.
