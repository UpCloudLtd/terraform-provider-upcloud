---
layout: "upcloud"
page_title: "UpCloud: resource_upcloud_router"
sidebar_current: "docs-upcloud-resource-upcloud-router"
description: |-
  Provides an UpCloud router
---

# upcloud_router

This resource represents a generated UpCloud router resource.  Routers can be used to connect multiple Private Networks. 
UpCloud Servers on any attached network can communicate directly with each other. 

## Example Usage

```hcl
resource "upcloud_router" "my_example_router" {
  name = "My Example Router"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to be assigned to the router

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `type` - The type assigned to the router.

* `attached_networks` - A list of network UUID that are attached through this router.

## Import

Existing UpCloud routers can be imported into the current Terraform state through the assigned UUID.

```hcl
terraform import upcloud_router.my_example_router 049d7ca2-757e-4fb1-a833-f87ee056547a
```