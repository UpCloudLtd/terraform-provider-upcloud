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

* `hostname` - (Required)
* `zone` - (Required)
* `firewall` - (Optional)
* `cpu` - (Optional)
* `mem` - (Optional)
* `template` - (Optional)
* `private_networking` - (Optional)
* `ipv4` - (Optional)
* `ipv6` - (Optional) 
* `user_data` - (Optional)
* `plan` - (Optional)
* `storage_devices` - (Required)
* `login` - (Optional)

The `storage_devices` block supports:

* `id` - 
* `address` - (Optional)
* `action` - (Required)
* `size` - (Optional)
* `tier` - (Optional)
* `title` - (Optional)
* `storage` - (Optional)
* `type` - (Optional)
* `backup_rule` - (Optional)

The `backup_rule` block supports:

* `interval` - (Required)
* `time` - (Required)
* `retention` - (Required)


The `login` block supports:

* `user` - (Required)
* `keys` - (Optional)
* `create_password` - (Optional) 
* `password_delivery` - (Optional)

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `title` - 
* `ipv4_address` -  
* `ipv4_address_private` - 
* `ipv6_address` -

## Import
