---
layout: "upcloud"
page_title: "UpCloud: datasource_upcloud_zone"
sidebar_current: "docs-upcloud-datasource-upcloud-zone"
description: |-
  Provides an UpCloud zone datasource
---

# upcloud_zone

Returns attributes of an available UpCloud zone.  All servers and storages can set their zone attributes to zone name returned by this datasource.

## Example Usage

The following example will return a zone.

```hcl
data "upcloud_zone" "my_zone" {
  name = "uk-lon1"
}
```

## Argument Reference

The following arguments can be supplied to the datasource.

* `name` - (Required) The UpCloud value that represents the zone. 

## Attributes Reference

* `description` - A real world value representing the zone, for example `London #1`.
* `public` - . Identifies the zone as either public `true` or private `false`.
