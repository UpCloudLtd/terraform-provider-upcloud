---
layout: "upcloud"
page_title: "UpCloud: resource_upcloud_tag"
sidebar_current: "docs-upcloud-resource-upcloud-tag"
description: |-
  Provides an UpCloud Tag
---

# upcloud_tag

This resource represents a UpCloud Tag resource.

## Example Usage

The following HCL example shows the creation of a Tag resource.

```hcl
    resource "upcloud_server" "my_example_server" {
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

    resource "upcloud_tag" "dev" {
    
      name = "development"
      description = "Represents the development environment"
      servers = [
        upcloud_server.my_example_server.id,

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