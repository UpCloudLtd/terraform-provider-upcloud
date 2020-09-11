---
layout: "upcloud"
page_title: "UpCloud: datasource_upcloud_tags"
sidebar_current: "docs-upcloud-datasource-upcloud-tags"
description: |-
  Provides an UpCloud tags datasource
---

# upcloud_tags

Returns a list of available UpCloud tags associated with the account.

## Example Usage

The following example will return a complete collection of tags used within the account.

```hcl
data "upcloud_tags" "all_tags" {}
``` 

## Attributes Reference
 
 The following attributes are returned to represent the collection of tags.
 
 * `tags` - A set of tags
 
 Each element of the tags set provides the following attributes
 
 * `name` - The value representing the tag
 
 * `description` - Free form text representing the meaning of the tag
 
 * `servers` - A collection of servers that have been assigned the tag