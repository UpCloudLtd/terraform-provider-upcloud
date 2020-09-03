---
layout: "upcloud"
page_title: "UpCloud: resource_upcloud_storage"
sidebar_current: "docs-upcloud-resource-upcloud-storage"
description: |-
  Provides an UpCloud Storage
---

# upcloud_storage

This resource represents a UpCloud Storage resource.

## Example Usage

The following HCL example shows the creation of a storage resource.

```hcl
    resource "upcloud_storage" "example_storage" {
      size  = 10
      tier  = "maxiops"
      title = "My data collection"
      zone  = "fi-hel1"
    }
```

The following HCL example shows the creation of a storage resource with the optional backup rule.
This storage resource will be backed up daily at 01:00 hours and each backup will be retained for 8 days.

```hcl
    resource "upcloud_storage" "example_storage_backup" {
      size  = 10
      tier  = "maxiops"
      title = "My data collection backup"
      zone  = "fi-hel1"
    
      backup_rule {
        interval  = "daily"
        time      = "0100"
        retention = 8
      }
    }
```

The following HCL example shows the creation of the storage resource with the optional import block.
This storage resource will have its content imported from an accessible website:

```hcl
    resource "upcloud_storage" "example_storage_backup" {
      size  = 10
      tier  = "maxiops"
      title = "My imported data"
      zone  = "fi-hel1"
    
      import {
        source = "http_import"
        source_location = "http://dl-cdn.alpinelinux.org/alpine/v3.12/releases/x86/alpine-standard-3.12.0-x86.iso"
      }
    }
```

The following HCL example shows the creation of the storage resource with the optional import block.
This storage resource will have its content imported from a local file:

```hcl
    resource "upcloud_storage" "example_storage_backup" {
      size  = 10
      tier  = "maxiops"
      title = "My imported data"
      zone  = "fi-hel1"
    
      import {
        source = "direct_upload"
        source_location = "/tmp/upload_image.img"
        source_hash = filesha256("/tmp/upload_image.img")
      }
    }
```

The following HCL example shows the creation of the storage resource with the optional clone block.
This storage resource will be cloned from the referenced storage ID. The reference storage should either not
be attached to a server or that server be stopped. If the storage to clone is not the specified size
the storage will be resized after cloning.

```hcl
    resource "upcloud_storage" "example_storage_clone" {
      size  = 20
      tier  = "maxiops"
      title = "My cloned data"
      zone  = "fi-hel1"
    
      clone {
        id = "01f936c9-38b2-4a10-b1fe-ad43d3078246"
      }
    }
```

## Argument Reference

The following arguments are supported:

* `size` - (Required) The size of the storage in gigabytes
* `tier` - (Optional) The storage tier to use
* `title` - (Required) A short, informative description
* `zone` - (Required) The zone in which the storage will be created
* `backup_rule` - (Optional) The criteria to backup the storage
* `import` - (Optional) Details of the external data to import
* `clone` - (Optional) Details of another storage device to clone

The `backup_rule` block supports:

* `interval` - (Required) The weekday when the backup is created
* `time` - (Required) The time of day when the backup is created
* `retention` - (Required) The number of days before a backup is automatically deleted

The `import` block supports:

* `source` - (Required) The source type (one of `direct_upload` or `http_import`).
* `source_location` - (Required) For `direct_upload` the path to a local file. For `http_import` an accessible URL.
* `source_hash` - (Optional) The hash of `source_location`. This is used to indicate that `source_location` has changed. It is not used for verification.

The `clone` block supports:

* `id` - (Required) The unique identifier of another storage device to clone.

## Import

Existing UpCloud storage can be imported into the current Terraform state through the assigned UUID.

```hcl
  terraform import upcloud_storage.example_storage 0128ae5a-91dd-4ebf-bd1e-304c47f2c652
```