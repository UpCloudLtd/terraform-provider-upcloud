---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "upcloud_gateway Resource - terraform-provider-upcloud"
subcategory: ""
description: |-
  Network gateways connect SDN Private Networks to external IP networks.
---

# upcloud_gateway (Resource)

Network gateways connect SDN Private Networks to external IP networks.

## Example Usage

```terraform
// Create router for the gateway
resource "upcloud_router" "this" {
  name = "gateway-example-router"
}

// Create network for the gateway
resource "upcloud_network" "this" {
  name = "gateway-example-net"
  zone = "pl-waw1"

  ip_network {
    address = "172.16.2.0/24"
    dhcp    = true
    family  = "IPv4"
  }

  router = upcloud_router.this.id
}

resource "upcloud_gateway" "this" {
  name     = "gateway-example-gw"
  zone     = "pl-waw1"
  features = ["nat"]

  router {
    id = upcloud_router.this.id
  }

  labels = {
    managed-by = "terraform"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `features` (Set of String) Features enabled for the gateway.
- `name` (String) Gateway name. Needs to be unique within the account.
- `router` (Block List, Min: 1, Max: 1) Attached Router from where traffic is routed towards the network gateway service. (see [below for nested schema](#nestedblock--router))
- `zone` (String) Zone in which the gateway will be hosted, e.g. `de-fra1`.

### Optional

- `configured_status` (String) The service configured status indicates the service's current intended status. Managed by the customer.
- `labels` (Map of String) Key-value pairs to classify the network gateway.

### Read-Only

- `id` (String) The ID of this resource.
- `operational_state` (String) The service operational state indicates the service's current operational, effective state. Managed by the system.

<a id="nestedblock--router"></a>
### Nested Schema for `router`

Required:

- `id` (String) ID of the router attached to the gateway.

