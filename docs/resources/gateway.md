---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "upcloud_gateway Resource - terraform-provider-upcloud"
subcategory: Network
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

  # UpCloud Network Gateway Service will add a static route to this router to ensure gateway networking is working as intended.
  # You need to ignore changes to it, otherwise TF will attempt to remove the static routes on subsequent applies
  lifecycle {
    ignore_changes = [static_route]
  }
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

### Required Attributes

- `features` (Set of String) Features enabled for the gateway. Note that VPN feature is currently in beta, for more details see https://upcloud.com/resources/docs/networking#nat-and-vpn-gateways.
- `name` (String) Gateway name. Needs to be unique within the account.
- `zone` (String) Zone in which the gateway will be hosted, e.g. `de-fra1`.

### Optional Attributes

- `configured_status` (String) The service configured status indicates the service's current intended status. Managed by the customer.
- `labels` (Map of String) User defined key-value pairs to classify the network gateway.
- `plan` (String) Gateway pricing plan.

### Blocks

- `address` (Block Set, Max: 1) IP addresses assigned to the gateway. (see [below for nested schema](#nestedblock--address))
- `router` (Block List, Min: 1, Max: 1) Attached Router from where traffic is routed towards the network gateway service. (see [below for nested schema](#nestedblock--router))

### Read-Only

- `addresses` (Set of Object, Deprecated) IP addresses assigned to the gateway. (see [below for nested schema](#nestedatt--addresses))
- `connections` (List of String) Names of connections attached to the gateway. Note that this field can have outdated information as connections are created by a separate resource. To make sure that you have the most recent data run 'terrafrom refresh'.
- `id` (String) The ID of this resource.
- `operational_state` (String) The service operational state indicates the service's current operational, effective state. Managed by the system.

<a id="nestedblock--address"></a>
### Nested Schema for `address`

Optional Attributes:

- `name` (String) Name of the IP address

Read-Only:

- `address` (String) IP addresss


<a id="nestedblock--router"></a>
### Nested Schema for `router`

Required Attributes:

- `id` (String) ID of the router attached to the gateway.


<a id="nestedatt--addresses"></a>
### Nested Schema for `addresses`

Read-Only:

- `address` (String)
- `name` (String)
