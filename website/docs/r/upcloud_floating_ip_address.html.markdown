---
layout: "upcloud"
page_title: "UpCloud: resource_upcloud_floating_ip_address"
sidebar_current: "docs-upcloud-resource-upcloud-floating-ip-address"
description: |-
  Provides an UpCloud Floating IP Address
---

# upcloud_floating_ip_address

This resource represents a UpCloud floating IP address resource.


## Example Usage

The following HCL example shows how to create a detached floating IP address.
```hcl
    resource "upcloud_floating_ip_address" "my_floating_address" {
      zone     = "fi-hel1"
    }
```

The following HCL example shows the creation of a floating IP address assigned to a server resource.

```hcl
    resource "upcloud_server" "my_server" {
      zone     = "fi-hel1"
      hostname = "mydebian.example.com"
      plan     = "1xCPU-2GB"
    
      storage_devices {
        action = "create"
        size   = 10
        tier   = "maxiops"
      }
    
      network_interface {
        type = "public"
      }
    
    }
    
    resource "upcloud_floating_ip_address" "my_new_floating_address" {
      mac_address = upcloud_server.my_server.network_interface[0].mac_address
    }
```

## Argument Reference

The following arguments are supported:

* `access` - (Optional) The IP address access type (one of `utility` or `public`)
* `family` - (Optional) The IP address family (one of `IPv4` or `IPv6`)
* `mac_address` - (Optional) MAC address of server interface to assign address to

## Attributes Reference

* `ip_address` - An UpCloud assigned IP Address
* `zone` - Zone of address, required when assigning a detached floating IP address.  Required when defining a detached floating IP address resource.

## Import

An existing UpCloud Floating IP address can be imported into the current Terraform state through the assigned IP Address.

```hcl
   terraform import upcloud_floating_ip_address.my_new_floating_address 94.237.114.205
```