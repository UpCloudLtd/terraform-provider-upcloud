---
layout: "upcloud"
page_title: "UpCloud: resource_upcloud_network"
sidebar_current: "docs-upcloud-resource-upcloud-network"
description: |-
  Provides an UpCloud network
---

# upcloud_network

This resource represents an SDN private network that cloud servers from the same zone can be attached to.

## Example Usage

```hcl
resource "upcloud_network" "example_network" {
  name = "example_private_net"
  zone = "nl-ams1"

  router = upcloud_router.example_router.id

  ip_network {
    address            = "10.0.0.0/24"
    dhcp               = true
    dhcp_default_route = false
    family  = "IPv4"
    gateway = "10.0.0.1"
  }
}

resource "upcloud_router" "example_router" {
  name = "example_router"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to be assigned to the network.
* `zone` - The UpCloud zone this network will reside in.
* `router` - A router to connect networks together.
* `ip_network` - A block containing the IP details of the network.

### ip_network

* `address` - (Required) The CIDR range of the network. Updating forces a new `network` resource.
* `dhcp` - (Required) Is DHCP enabled?
* `dhcp_default_route` - Is the `gateway` the DHCP default route?
* `dhcp_dns` - A set of the DNS servers given by DHCP.
* `family` - (Required) The IP address family (one of `IPv4` or `IPv6`)
* `gateway` - The gateway address given by DHCP.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `type` - The type assigned to the network.
* `servers` - A set of attached servers

### servers

* `id` - The unique identifier of the server.
* `title` - The short description of the server.

## Import

Existing UpCloud networks can be imported into the current Terraform state through the assigned UUID.

```hcl
terraform import upcloud_network.my_example_network 03e44422-07b8-4798-a597-c8eab1fa64df
```