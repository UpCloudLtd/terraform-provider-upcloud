---
layout: "upcloud"
page_title: "UpCloud: upcloud_networks"
sidebar_current: "docs-upcloud-datasource-networks"
description: |-
  Get information on available UpCloud networks.
---

# upcloud_networks

Use this data source to get the available UpCloud [networks][1].

## Example Usage

The following example will return all available networks:

```hcl
data "upcloud_networks" "upcloud" {}
```

The following example will return all available networks within a zone:

```hcl
data "upcloud_networks" "upcloud_by_zone" {
  zone = "fi-hel1"
}
```

The following example will return all available networks filtered by a
regular expression on the name of the network:

```hcl
data "upcloud_networks" "upcloud_by_zone" {
  filter_name = "^Public.*"
}
```

## Argument Reference

The following arguments are supported:

* `zone` - (Optional) Limit returned networks to this UpCloud zone.
* `filter_name` - (Optional) A regular-expression allowing filtering of networks by name.

## Attributes Reference

* `networks` - A set of the available networks.

### networks

* `ip_network` - A set of the IP subnets within the network.
* `name` - The name of the network.
* `type` - The type of network (one of `public`, `utility`, `private`)
* `id` - The unique identifier of the network.
* `zone` - The zone in which this network resides.
* `servers` - A set of the servers joined to this network.

### ip_network

* `address` - The CIRD range of the subnet.
* `dhcp` - Is DHCP enabled?
* `dhcp_default_route` - Is the `gateway` the DHCP default route?
* `dhcp_dns` - A set of the DNS servers given by DHCP.
* `family` - The IP address familty (one of `IPv4` or `IPv6`)
* `gateway` - The gateway address given by DHCP.

### servers

* `id` - The unique identifier of the server.
* `title` - The short description of the server.

[1]: https://upcloud.com/products/software-defined-networking/
