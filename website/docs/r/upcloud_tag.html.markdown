---
layout: "upcloud"
page_title: "UpCloud: resource_upcloud_tag"
sidebar_current: "docs-upcloud-resource-upcloud-tag"
description: |-
  Provides an UpCloud Tag
---

# upcloud_tag

This resource represents a UpCloud Tag resource. Tags are currently not fully supported for sub accounts.

## Example Usage

The following HCL example shows the creation of a Tag resource.

```hcl
resource "upcloud_server" "example" {
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"

  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
    size    = 25
  }

  network_interface {
    type = "public"
  }
}

resource "upcloud_tag" "dev" {
  name = "development"
  description = "Represents the development environment"
  servers = [
    upcloud_server.example.id,
  ]
}
```


## Argument Reference

The following arguments are supported:

 * `name` - The value representing the tag

 * `description` - Freeform comment string for the host

 * `servers` - A collection of servers that have been assigned the tag.

## Import

An Existing UpCloud Tag can be imported into the current Terraform state through the tag name.

```hcl
  terraform import upcloud_tag.example_tag dev
```
