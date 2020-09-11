---
layout: "upcloud"
page_title: "UpCloud: datasource_upcloud_ip_addresses"
sidebar_current: "docs-upcloud-datasource-upcloud-ip-addresses"
description: |-
  Provides an UpCloud IP Addresses datasource
---

# upcloud_ip_addresses

Returns a set of IP Addresses that are associated with the UpCloud account.  

## Example Usage

The following example will return a complete set of IP addresses associated with the account.

```hcl
data "upcloud_ip_addresses" "all_ip_addresses" {}
```


## Argument Reference

The UpCloud IP Addresses datasource does not accept any arguments.

## Attributes Reference

* access - Is address for utility or public network
* address - An UpCloud assigned IP Address
* family - IP address family
* part_of_plan - Is the address a part of a plan
* ptr_record - A reverse DNS record entry
* server - The unique identifier for a server
* mac - MAC address of server interface to assign address to
* floating - Does the IP Address represents a floating IP Address 
