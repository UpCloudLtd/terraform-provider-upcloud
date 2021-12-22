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
