---
layout: "upcloud"
page_title: "UpCloud: resource_upcloud_server"
sidebar_current: "docs-upcloud-resource-upcloud-server"
description: |-
  Provides an UpCloud service
---

# upcloud_server

This resource represents a generated UpCloud server resource.

## Example Usage

```hcl
resource "upcloud_server" "example" {
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"

  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
    size    = 25

    backup_rule {
      interval  = "daily"
      time      = "0100"
      retention = 8
    }
  }

  network_interface {
    type = "public"
  }

  login {
    user = "myusername"

    keys = [
      "<YOUR SSH PUBLIC KEY>",
    ]

    create_password   = true
    password_delivery = "sms"
  }
}
```

## Argument Reference

The following arguments are supported:

* `hostname` - (Required) A valid domain name, e.g. host.example.com. The maximum length is 128 characters.
* `zone` - (Required) - The zone in which the server will be hosted, e.g. fi-hel1. See [Zones API](https://developers.upcloud.com/1.3/5-zones/)
* `network_interface` - (Required) One or more blocks describing the network interfaces of the server. Any changes to these blocks will force the creation of a new server resource.
* `firewall` - (Optional) Are firewall rules active for the server
* `metadata` - (Optional) Is the metadata service active for the server
* `cpu` - (Optional) The number of CPU for the server
* `mem` - (Optional) The size of memory for the server (in megabytes)
* `template` - (Optional) The template to use for the server's main storage device
* `user_data` - (Optional) Defines URL for a server setup script, or the script body itself
* `plan` - (Optional) The pricing plan used for the server
* `storage_devices` - (Optional) A list of storage devices associated with the server
* `login` - (Optional) Configure access credentials to the server

The `storage_devices` block supports:

* `storage` - (Required) A valid storage UUID
* `address` - (Optional) The device address the storage will be attached to. Specify only the bus name (ide/scsi/virtio) to auto-select next available address from that bus.
* `type` - (Optional) The device type the storage will be attached as. See [Storage types](https://developers.upcloud.com/1.3/9-storages/).

The `template` block supports:

* `storage` - (Required) A valid storage UUID or exact template name
* `address` - (Optional) The device address the storage will be attached to. Specify only the bus name (ide/scsi/virtio) to auto-select next available address from that bus.
* `title` - (Optional) A short, informative description for the storage device
* `size`- (Optional) The size of the storage in gigabytes
* `backup_rule` - (Optional) The criteria to backup the storage

The `login` block supports:

* `user` - (Required) Username to be create to access the server
* `keys` - (Optional) A list of ssh keys to access the server
* `create_password` - (Optional) Indicates a password should be create to allow access
* `password_delivery` - (Optional) The delivery method for the server’s root password

The `network_interface` block supports:

* `type` - (Required) The type of network interface (one of `private`, `public` or `utility`).
* `ip_address_family` - (Optional) The IP address type of this interface (one of `IPv4` or `IPv6`).
* `network` - (Optional) The unique ID of a network to attach this interface to. Only supported for `private` interfaces.
* `source_ip_filtering` - (Optional) `true` if source IPs should be filtered. Only supported for `private` interfaces.
* `bootable` - (Optional) `true` if this interface should be used for network booting. Only supported for `private` interfaces.

The `backup_rule` block supports:

* `interval` - (Required) The weekday when the backup is created
* `time` - (Required) The time of day when the backup is created
* `retention` - (Required) The number of days before a backup is automatically deleted

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of this resource.
* `title` - A short, informational description.

In addition to the arguments listed above, the following `template` block attributes are exported:

* `id` - The unique identifier for the storage
* `tier` - The storage tier to use

In addition to the arguments listed above, the following `network_interface` block attributes are exported:

* `ip_address` - The assigned IP address.
* `ip_address_floating` - `true` if a floating IP address is attached.
* `mac_address` - The assigned MAC address.

## Import
