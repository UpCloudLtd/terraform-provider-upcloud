---
layout: "upcloud"
page_title: "UpCloud: datasource_upcloud_zones"
sidebar_current: "docs-upcloud-datasource-upcloud-zones"
description: |-
  Provides an UpCloud zones datasource
---

# upcloud_zones

Returns a list of available UpCloud zones.  All servers and storages must set their zone attributes to one of the zone ids returned by this datasource.

## Example Usage

The following example will return a complete list of public and private zones.

```hcl
data "upcloud_zones" "all_zones" {}
```

The following example will return a list of only public zones.

```hcl
data "upcloud_zones" "all_public_zones" {
  filter_type = "public"
}
```

The following example will return a list of only private zones.

```hcl
data "upcloud_zones" "all_private_zones" {
  filter_type = "private"
}
```

## Argument Reference

The following arguments can be supplied to the datasource.

* `filter_type` - (Optional) Allows the zone_ids to be filtered by type, (public, private or all). Defaults to `all` 

## Attributes Reference

* `zone_ids` - A Collection of ids representing the available zones within UpCloud. 
