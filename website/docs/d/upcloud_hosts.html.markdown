---
layout: "upcloud"
page_title: "UpCloud: datasource_upcloud_zones"
sidebar_current: "docs-upcloud-datasource-upcloud-zones"
description: |-
  Provides an UpCloud hosts datasource
---

# upcloud_hosts

Returns a list of available UpCloud hosts.  A host identifies the host server that virtual machines are run on.
Only hosts on private cloud to which the calling account has access to are available through this resource.

## Example Usage

The following example will return a list of hosts.

```hcl
data "upcloud_hosts" "my_hosts" {}
``` 

## Attributes Reference

* `hosts` - A list of hosts within your private cloud that can be accessed by the account

### hosts

Each entry in the `hosts` attribute represents an individual host from your account.  The following attributes are available. 

Each `hosts` entry provides the following:
 
* `host_id` - The unique ID representing the host

* `description` - Freeform comment string for the host

* `zone` -  The zone ID where the hosts is located.