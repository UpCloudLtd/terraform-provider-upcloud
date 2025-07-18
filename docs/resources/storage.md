---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "upcloud_storage Resource - terraform-provider-upcloud"
subcategory: Storage
description: |-
  Manages UpCloud Block Storage https://upcloud.com/products/block-storage devices.
---

# upcloud_storage (Resource)

Manages UpCloud [Block Storage](https://upcloud.com/products/block-storage) devices.

## Example Usage

```terraform
# Storage resource.
resource "upcloud_storage" "example_storage" {
  size  = 10
  tier  = "maxiops"
  title = "My data collection"
  zone  = "fi-hel1"
}

# Storage resource with the optional backup rule. 
# This storage resource will be backed up daily at 01:00 hours and each backup will be retained for 8 days.
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

# Storage resource with the optional import block. 
# This storage resource will have its content imported from an accessible website:
resource "upcloud_storage" "example_storage_backup" {
  size  = 10
  tier  = "maxiops"
  title = "My imported data"
  zone  = "fi-hel1"

  import {
    source          = "http_import"
    source_location = "http://dl-cdn.alpinelinux.org/alpine/v3.12/releases/x86/alpine-standard-3.12.0-x86.iso"
  }
}

# Storage resource with the optional import block. 
# This storage resource will have its content imported from a local file:
resource "upcloud_storage" "example_storage_backup" {
  size  = 10
  tier  = "maxiops"
  title = "My imported data"
  zone  = "fi-hel1"

  import {
    source          = "direct_upload"
    source_location = "/tmp/upload_image.img"
    source_hash     = filesha256("/tmp/upload_image.img")
  }
}

# Storage resource with the optional clone block. 
# This storage resource will be cloned from the referenced storage ID. 
# The reference storage should either not be attached to a server or that server be stopped. 
# If the storage to clone is not the specified size the storage will be resized after cloning.
resource "upcloud_storage" "example_storage_clone" {
  size  = 20
  tier  = "maxiops"
  title = "My cloned data"
  zone  = "fi-hel1"

  clone {
    id = "01f936c9-38b2-4a10-b1fe-ad43d3078246"
  }
}

# Storage resource with the creation of a server resource which will attach the created storage resource.
resource "upcloud_storage" "example_storage" {
  size  = 20
  tier  = "maxiops"
  title = "My storage"
  zone  = "fi-hel1"
}

resource "upcloud_server" "example_server" {
  zone     = "fi-hel1"
  plan     = "1xCPU-1GB"
  hostname = "terraform.example.tld"

  network_interface {
    type = "public"
  }

  storage_devices {
    storage = upcloud_storage.example_storage[0].id
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required Attributes

- `size` (Number) The size of the storage in gigabytes.
- `title` (String) The title of the storage.
- `zone` (String) The zone the storage is in, e.g. `de-fra1`. You can list available zones with `upctl zone list`.

### Optional Attributes

- `delete_autoresize_backup` (Boolean) If set to true, the backup taken before the partition and filesystem resize attempt will be deleted immediately after success.
- `encrypt` (Boolean) Sets if the storage is encrypted at rest.
- `filesystem_autoresize` (Boolean) If set to true, provider will attempt to resize partition and filesystem when the size of the storage changes. Please note that before the resize attempt is made, backup of the storage will be taken. If the resize attempt fails, the backup will be used to restore the storage and then deleted. If the resize attempt succeeds, backup will be kept (unless `delete_autoresize_backup` option is set to true).
				Taking and keeping backups incure costs.
- `labels` (Map of String) User defined key-value pairs to classify the storage.
- `tier` (String) The tier of the storage.

### Blocks

- `backup_rule` (Block List) The criteria to backup the storage.

    Please keep in mind that it's not possible to have a storage with `backup_rule` attached to a server with `simple_backup` specified. Such configurations will throw errors during execution.

    Also, due to how UpCloud API works with simple backups and how Terraform orders the update operations, it is advised to never switch between `simple_backup` on the server and individual storages `backup_rules` in one apply. If you want to switch from using server simple backup to per-storage defined backup rules,  please first remove `simple_backup` block from a server, run `terraform apply`, then add `backup_rule` to desired storages and run `terraform apply` again. (see [below for nested schema](#nestedblock--backup_rule))
- `clone` (Block Set) Block defining another storage/template to clone to storage. (see [below for nested schema](#nestedblock--clone))
- `import` (Block Set) Block defining external data to import to storage (see [below for nested schema](#nestedblock--import))

### Read-Only

- `id` (String) UUID of the storage.
- `system_labels` (Map of String) System defined key-value pairs to classify the storage. The keys of system defined labels are prefixed with underscore and can not be modified by the user.
- `type` (String) The type of the storage.

<a id="nestedblock--backup_rule"></a>
### Nested Schema for `backup_rule`

Required Attributes:

- `interval` (String) The weekday when the backup is created
- `retention` (Number) The number of days before a backup is automatically deleted
- `time` (String) The time of day (UTC) when the backup is created


<a id="nestedblock--clone"></a>
### Nested Schema for `clone`

Required Attributes:

- `id` (String) The unique identifier of the storage/template to clone.


<a id="nestedblock--import"></a>
### Nested Schema for `import`

Required Attributes:

- `source` (String) The mode of the import task. One of `http_import` or `direct_upload`.
- `source_location` (String) The location of the file to import. For `http_import` an accessible URL. For `direct_upload` a local file. When direct uploading a compressed image, `Content-Type` header of the PUT request is set automatically based on the file extension (`.gz` or `.xz`, case-insensitive).

Optional Attributes:

- `source_hash` (String) SHA256 hash of the source content. This hash is used to verify the integrity of the imported data by comparing it to `sha256sum` after the import has completed. Possible filename is automatically removed from the hash before comparison.

Read-Only:

- `sha256sum` (String) sha256 sum of the imported data
- `written_bytes` (Number) Number of bytes imported

## Import

Import is supported using the following syntax:

```shell
terraform import upcloud_storage.example_storage 0128ae5a-91dd-4ebf-bd1e-304c47f2c652
```
