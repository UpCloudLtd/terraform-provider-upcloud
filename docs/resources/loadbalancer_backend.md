---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "upcloud_loadbalancer_backend Resource - terraform-provider-upcloud"
subcategory: Load Balancer
description: |-
  This resource represents load balancer backend service.
---

# upcloud_loadbalancer_backend (Resource)

This resource represents load balancer backend service.

## Example Usage

```terraform
variable "lb_zone" {
  type    = string
  default = "fi-hel2"
}

resource "upcloud_network" "lb_network" {
  name = "lb-test-net"
  zone = var.lb_zone
  ip_network {
    address = "10.0.0.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_loadbalancer" "lb" {
  configured_status = "started"
  name              = "lb-test"
  plan              = "development"
  zone              = var.lb_zone
  network           = resource.upcloud_network.lb_network.id
}

resource "upcloud_loadbalancer_backend" "lb_be_1" {
  loadbalancer = resource.upcloud_loadbalancer.lb.id
  name         = "lb-be-1-test"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required Attributes

- `loadbalancer` (String) UUID of the load balancer to which the backend is connected.
- `name` (String) The name of the backend. Must be unique within the load balancer service.

### Optional Attributes

- `resolver_name` (String) Domain name resolver used with dynamic type members.

### Blocks

- `properties` (Block List) Backend properties. Properties can be set back to defaults by defining an empty `properties {}` block. For `terraform import`, an empty or non-empty block is also required. (see [below for nested schema](#nestedblock--properties))

### Read-Only

- `id` (String) ID of the backend. ID is in `{load balancer UUID}/{backend name}` format.
- `members` (List of String) Backend member server UUIDs. Members receive traffic dispatched from the frontends.
- `tls_configs` (List of String) Set of TLS config names.

<a id="nestedblock--properties"></a>
### Nested Schema for `properties`

Optional Attributes:

- `health_check_expected_status` (Number) Expected HTTP status code returned by the customer application to mark server as healthy. Ignored for `tcp` `health_check_type`.
- `health_check_fall` (Number) Sets how many failed health checks are allowed until the backend member is taken off from the rotation.
- `health_check_interval` (Number) Interval between health checks in seconds.
- `health_check_rise` (Number) Sets how many successful health checks are required to put the backend member back into rotation.
- `health_check_tls_verify` (Boolean) Enables certificate verification with the system CA certificate bundle. Works with https scheme in health_check_url, otherwise ignored.
- `health_check_type` (String) Health check type.
- `health_check_url` (String) Target path for health check HTTP GET requests. Ignored for `tcp` `health_check_type`.
- `http2_enabled` (Boolean) Allow HTTP/2 connections to backend members by utilizing ALPN extension of TLS protocol, therefore it can only be enabled when tls_enabled is set to true. Note: members should support HTTP/2 for this setting to work.
- `outbound_proxy_protocol` (String) Enable outbound proxy protocol by setting the desired version. Defaults to empty string. Empty string disables proxy protocol.
- `sticky_session_cookie_name` (String) Sets sticky session cookie name. Empty string disables sticky session.
- `timeout_server` (Number) Backend server timeout in seconds.
- `timeout_tunnel` (Number) Maximum inactivity time on the client and server side for tunnels in seconds.
- `tls_enabled` (Boolean) Enables TLS connection from the load balancer to backend servers.
- `tls_use_system_ca` (Boolean) If enabled, then the system CA certificate bundle will be used for the certificate verification.
- `tls_verify` (Boolean) Enables backend servers certificate verification. Please make sure that TLS config with the certificate bundle of type authority attached to the backend or `tls_use_system_ca` enabled. Note: `tls_verify` has preference over `health_check_tls_verify` when `tls_enabled` in true.
