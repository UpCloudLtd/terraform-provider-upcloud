---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "upcloud_managed_database_mysql_sessions Data Source - terraform-provider-upcloud"
subcategory: ""
description: |-
  Current sessions of a MySQL managed database
---

# upcloud_managed_database_mysql_sessions (Data Source)

Current sessions of a MySQL managed database

## Example Usage

```terraform
# Use data source to gather a list of the active sessions for a Managed MySQL Database

# Create a Managed MySQL resource
resource "upcloud_managed_database_mysql" "example" {
  name = "mysql-example1"
  plan = "1x1xCPU-2GB-25GB"
  zone = "fi-hel1"
}

# Read the active sessions of the newly created service
data "upcloud_managed_database_mysql_sessions" "example" {
  service = upcloud_managed_database_mysql.example.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `service` (String) Service's UUID for which these sessions belongs to

### Optional

- `limit` (Number) Number of entries to receive at most.
- `offset` (Number) Offset for retrieved results based on sort order.
- `order` (String) Order by session field and sort retrieved results. Limited variables can be used for ordering.

### Read-Only

- `id` (String) The ID of this resource.
- `sessions` (Block Set) Current sessions (see [below for nested schema](#nestedblock--sessions))

<a id="nestedblock--sessions"></a>
### Nested Schema for `sessions`

Read-Only:

- `application_name` (String) Name of the application that is connected to this service.
- `client_addr` (String) IP address of the client connected to this service.
- `datname` (String) Name of the database this service is connected to.
- `id` (String) Process ID of this service.
- `query` (String) Text of this service's most recent query. If state is active this field shows the currently executing query. In all other states, it shows an empty string.
- `query_duration` (String) The active query current duration.
- `state` (String) Current overall state of this service: active: The service is executing a query, idle: The service is waiting for a new client command.
- `usename` (String) Name of the user logged into this service.

