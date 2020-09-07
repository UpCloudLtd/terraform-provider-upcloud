---
layout: "upcloud"
page_title: "UpCloud: resource_upcloud_server"
sidebar_current: "docs-upcloud-resource-upcloud-server"
description: |-
  Provides an UpCloud service
---

# upcloud_server

This resource represents a generated resource in the new Terraform Provider.  You will need to declare one for each backend entity you wish to expose and manage in the newly generated Terraform Provider. 

## Example Usage

```hcl
resource "upcloud_server" "test" {
  zone     = "fi-hel1"
  hostname = "ubuntu.example.tld"

  cpu = "2"
  mem = "1024"

  # Login details
  login {
    user = "myusername"

    keys = [
      "<YOUR SSH PUBLIC KEY>",
    ]

    create_password   = true
    password_delivery = "sms"
  }

  storage_devices {
    # You can use both storage template names and UUIDs
    size    = 50
    action  = "clone"
    tier    = "maxiops"
    storage = "Ubuntu Server 16.04 LTS (Xenial Xerus)"

    backup_rule = {
      interval  = "daily"
      time      = "0100"
      retention = 8
    }
  }

}

```

## Argument Reference

The following arguments are supported:

* `hostname` - (Required) A valid domain name, e.g. host.example.com. The maximum length is 128 characters.
* `zone` - (Required) - The zone in which the server will be hosted, e.g. fi-hel1. See [Zones API](https://developers.upcloud.com/1.3/5-zones/)
* `firewall` - (Optional) Are firewall rules active for the server
* `cpu` - (Optional) The number of CPU for the server
* `mem` - (Optional) The size of memory for the server
* `template` - (Optional) The template to use during creation
* `user_data` - (Optional) Defines URL for a server setup script, or the script body itself
* `plan` - (Optional) The pricing plan used for the server
* `storage_devices` - (Required) A list of storage devices associated with the server
* `login` - (Optional) Configure access credentials to the server
* `network_interface` - (Required) One or more blocks describing the network interfaces of the server. Any changes to these blocks will force the creation of a new server resource.

The `storage_devices` block supports:

* `address` - (Optional) An UpCloud assigned IP Address
* `action` - (Required) The method used to create or attach the specified storage. Valid values are `create`, `clone` or `attach`.
* `size` - (Optional) The size of the storage in gigabytes. Required for the `create` action.
* `tier` - (Optional) The storage tier to use
* `title` - (Optional) A short, informative description
* `storage` - (Optional) A valid storage UUID. Applicable only if action is `attach` or `clone`.

* `type` - (Optional) The device type the storage will be attached as. See [Storage types](https://developers.upcloud.com/1.3/9-storages/).
* `backup_rule` - (Optional) The criteria to backup the storage

The `backup_rule` block supports:

* `interval` - (Required) The weekday when the backup is created
* `time` - (Required) The time of day when the backup is created
* `retention` - (Required) The number of days before a backup is automatically deleted

The `login` block supports:

* `user` - (Required) Username to be create to access the server
* `keys` - (Optional) A list of ssh keys to access the server
* `create_password` - (Optional) Indicates a password should be create to allow access
* `password_delivery` - (Optional) The delivery method for the server’s root password

The `network_interface` block supports:

* `ip_address_family` - (Optional) The IP address type of this interface (one of `IPv4` or `IPv6`).
* `type` - (Required) The type of network interface (one of `private`, `public` or `utility`).
* `network` - (Optional) The unique ID of a network to attach this interface to. Only supported for `private` interfaces.
* `source_ip_filtering` - (Optional) `true` if source IPs should be filtered. Only supported for `private` interfaces.
* `bootable` - (Optional) `true` if this interface should be used for network booting. Only supported for `private` interfaces.

In addition to the arguments listed above, the following attributes are exported:

* `title` - A short, informational description.

The `network_interface` block also exports the following additional attributes:

* `ip_address` - The assigned IP address.
* `ip_address_floating` - `true` if a floating IP address is attached.
* `mac_address` - The assigned MAC address.

## Import