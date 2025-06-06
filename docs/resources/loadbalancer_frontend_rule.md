---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "upcloud_loadbalancer_frontend_rule Resource - terraform-provider-upcloud"
subcategory: Load Balancer
description: |-
  This resource represents load balancer frontend rule.
---

# upcloud_loadbalancer_frontend_rule (Resource)

This resource represents load balancer frontend rule.

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

resource "upcloud_loadbalancer_frontend_rule" "lb_fe_1_r1" {
  frontend = resource.upcloud_loadbalancer_frontend.lb_fe_1.id
  name     = "lb-fe-1-r1-test"
  priority = 10

  matchers {
    src_ip {
      value = "192.168.0.0/24"
    }
  }

  actions {
    use_backend {
      backend_name = resource.upcloud_loadbalancer_backend.lb_be_1.name
    }
  }
}

resource "upcloud_loadbalancer_frontend" "lb_fe_1" {
  loadbalancer         = resource.upcloud_loadbalancer.lb.id
  name                 = "lb-fe-1-test"
  mode                 = "http"
  port                 = 8080
  default_backend_name = resource.upcloud_loadbalancer_backend.lb_be_1.name
}

resource "upcloud_loadbalancer" "lb" {
  configured_status = "started"
  name              = "lb-test"
  plan              = "development"
  zone              = var.lb_zone

  networks {
    type   = "public"
    family = "IPv4"
    name   = "public"
  }

  networks {
    type    = "private"
    family  = "IPv4"
    name    = "private"
    network = resource.upcloud_network.lb_network.id
  }
}

resource "upcloud_loadbalancer_backend" "lb_be_1" {
  loadbalancer = resource.upcloud_loadbalancer.lb.id
  name         = "lb-be-1-test"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required Attributes

- `frontend` (String) ID of the load balancer frontend to which the frontend rule is connected.
- `name` (String) The name of the frontend rule. Must be unique within the frontend.
- `priority` (Number) Rule with the higher priority goes first. Rules with the same priority processed in alphabetical order.

### Optional Attributes

- `matching_condition` (String) Defines boolean operator used to combine multiple matchers. Defaults to `and`.

### Blocks

- `actions` (Block List) Rule actions. (see [below for nested schema](#nestedblock--actions))
- `matchers` (Block List) Set of rule matchers. If rule doesn't have matchers, then action applies to all incoming requests. (see [below for nested schema](#nestedblock--matchers))

### Read-Only

- `id` (String) ID of the frontend rule. ID is in `{load balancer UUID}/{frontend name}/{frontend rule name}` format.

<a id="nestedblock--actions"></a>
### Nested Schema for `actions`

Blocks:

- `http_redirect` (Block List) Redirects HTTP requests to specified location or URL scheme. Only either location or scheme can be defined at a time. (see [below for nested schema](#nestedblock--actions--http_redirect))
- `http_return` (Block List) Returns HTTP response with specified HTTP status. (see [below for nested schema](#nestedblock--actions--http_return))
- `set_forwarded_headers` (Block List) Adds 'X-Forwarded-For / -Proto / -Port' headers in your forwarded requests (see [below for nested schema](#nestedblock--actions--set_forwarded_headers))
- `set_request_header` (Block List) Set request header (see [below for nested schema](#nestedblock--actions--set_request_header))
- `set_response_header` (Block List) Set response header (see [below for nested schema](#nestedblock--actions--set_response_header))
- `tcp_reject` (Block List) Terminates a connection. (see [below for nested schema](#nestedblock--actions--tcp_reject))
- `use_backend` (Block List) Routes traffic to specified `backend`. (see [below for nested schema](#nestedblock--actions--use_backend))

<a id="nestedblock--actions--http_redirect"></a>
### Nested Schema for `actions.http_redirect`

Optional Attributes:

- `location` (String) Target location.
- `scheme` (String) Target scheme.
- `status` (Number) HTTP status code.


<a id="nestedblock--actions--http_return"></a>
### Nested Schema for `actions.http_return`

Required Attributes:

- `content_type` (String) Content type.
- `payload` (String) The payload.
- `status` (Number) HTTP status code.


<a id="nestedblock--actions--set_forwarded_headers"></a>
### Nested Schema for `actions.set_forwarded_headers`

Optional Attributes:

- `active` (Boolean)


<a id="nestedblock--actions--set_request_header"></a>
### Nested Schema for `actions.set_request_header`

Required Attributes:

- `header` (String) Header name.

Optional Attributes:

- `value` (String) Header value.


<a id="nestedblock--actions--set_response_header"></a>
### Nested Schema for `actions.set_response_header`

Required Attributes:

- `header` (String) Header name.

Optional Attributes:

- `value` (String) Header value.


<a id="nestedblock--actions--tcp_reject"></a>
### Nested Schema for `actions.tcp_reject`

Optional Attributes:

- `active` (Boolean) Indicates if the rule is active.


<a id="nestedblock--actions--use_backend"></a>
### Nested Schema for `actions.use_backend`

Required Attributes:

- `backend_name` (String) The name of the backend where traffic will be routed.



<a id="nestedblock--matchers"></a>
### Nested Schema for `matchers`

Blocks:

- `body_size` (Block List) Matches by HTTP request body size. (see [below for nested schema](#nestedblock--matchers--body_size))
- `body_size_range` (Block List) Matches by range of HTTP request body sizes. (see [below for nested schema](#nestedblock--matchers--body_size_range))
- `cookie` (Block List) Matches by HTTP cookie value. Cookie name must be provided. (see [below for nested schema](#nestedblock--matchers--cookie))
- `header` (Block List, Deprecated) Matches by HTTP header value. Header name must be provided. (see [below for nested schema](#nestedblock--matchers--header))
- `host` (Block List) Matches by hostname. Header extracted from HTTP Headers or from TLS certificate in case of secured connection. (see [below for nested schema](#nestedblock--matchers--host))
- `http_method` (Block List) Matches by HTTP method. (see [below for nested schema](#nestedblock--matchers--http_method))
- `http_status` (Block List) Matches by HTTP status. (see [below for nested schema](#nestedblock--matchers--http_status))
- `http_status_range` (Block List) Matches by range of HTTP statuses. (see [below for nested schema](#nestedblock--matchers--http_status_range))
- `num_members_up` (Block List) Matches by number of healthy backend members. (see [below for nested schema](#nestedblock--matchers--num_members_up))
- `path` (Block List) Matches by URL path. (see [below for nested schema](#nestedblock--matchers--path))
- `request_header` (Block List) Matches by HTTP request header value. Header name must be provided. (see [below for nested schema](#nestedblock--matchers--request_header))
- `response_header` (Block List) Matches by HTTP response header value. Header name must be provided. (see [below for nested schema](#nestedblock--matchers--response_header))
- `src_ip` (Block List) Matches by source IP address. (see [below for nested schema](#nestedblock--matchers--src_ip))
- `src_port` (Block List) Matches by source port number. (see [below for nested schema](#nestedblock--matchers--src_port))
- `src_port_range` (Block List) Matches by range of source port numbers. (see [below for nested schema](#nestedblock--matchers--src_port_range))
- `url` (Block List) Matches by URL without schema, e.g. `example.com/dashboard`. (see [below for nested schema](#nestedblock--matchers--url))
- `url_param` (Block List) Matches by URL query parameter value. Query parameter name must be provided (see [below for nested schema](#nestedblock--matchers--url_param))
- `url_query` (Block List) Matches by URL query string. (see [below for nested schema](#nestedblock--matchers--url_query))

<a id="nestedblock--matchers--body_size"></a>
### Nested Schema for `matchers.body_size`

Required Attributes:

- `method` (String) Match method (`equal`, `greater`, `greater_or_equal`, `less`, `less_or_equal`).
- `value` (Number) Integer value.

Optional Attributes:

- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.


<a id="nestedblock--matchers--body_size_range"></a>
### Nested Schema for `matchers.body_size_range`

Required Attributes:

- `range_end` (Number) Integer value.
- `range_start` (Number) Integer value.

Optional Attributes:

- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.


<a id="nestedblock--matchers--cookie"></a>
### Nested Schema for `matchers.cookie`

Required Attributes:

- `method` (String) Match method (`exact`, `substring`, `regexp`, `starts`, `ends`, `domain`, `ip`, `exists`). Matcher with `exists` and `ip` methods must be used without `value` and `ignore_case` fields.
- `name` (String) Name of the argument.

Optional Attributes:

- `ignore_case` (Boolean) Defines if case should be ignored. Defaults to `false`.
- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.
- `value` (String) String value.


<a id="nestedblock--matchers--header"></a>
### Nested Schema for `matchers.header`

Required Attributes:

- `method` (String) Match method (`exact`, `substring`, `regexp`, `starts`, `ends`, `domain`, `ip`, `exists`). Matcher with `exists` and `ip` methods must be used without `value` and `ignore_case` fields.
- `name` (String) Name of the argument.

Optional Attributes:

- `ignore_case` (Boolean) Defines if case should be ignored. Defaults to `false`.
- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.
- `value` (String) String value.


<a id="nestedblock--matchers--host"></a>
### Nested Schema for `matchers.host`

Required Attributes:

- `value` (String) String value.

Optional Attributes:

- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.


<a id="nestedblock--matchers--http_method"></a>
### Nested Schema for `matchers.http_method`

Required Attributes:

- `value` (String) String value (`GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`).

Optional Attributes:

- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.


<a id="nestedblock--matchers--http_status"></a>
### Nested Schema for `matchers.http_status`

Required Attributes:

- `method` (String) Match method (`equal`, `greater`, `greater_or_equal`, `less`, `less_or_equal`).
- `value` (Number) Integer value.

Optional Attributes:

- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.


<a id="nestedblock--matchers--http_status_range"></a>
### Nested Schema for `matchers.http_status_range`

Required Attributes:

- `range_end` (Number) Integer value.
- `range_start` (Number) Integer value.

Optional Attributes:

- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.


<a id="nestedblock--matchers--num_members_up"></a>
### Nested Schema for `matchers.num_members_up`

Required Attributes:

- `backend_name` (String) The name of the `backend`.
- `method` (String) Match method (`equal`, `greater`, `greater_or_equal`, `less`, `less_or_equal`).
- `value` (Number) Integer value.

Optional Attributes:

- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.


<a id="nestedblock--matchers--path"></a>
### Nested Schema for `matchers.path`

Required Attributes:

- `method` (String) Match method (`exact`, `substring`, `regexp`, `starts`, `ends`, `domain`, `ip`, `exists`). Matcher with `exists` and `ip` methods must be used without `value` and `ignore_case` fields.

Optional Attributes:

- `ignore_case` (Boolean) Defines if case should be ignored. Defaults to `false`.
- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.
- `value` (String) String value.


<a id="nestedblock--matchers--request_header"></a>
### Nested Schema for `matchers.request_header`

Required Attributes:

- `method` (String) Match method (`exact`, `substring`, `regexp`, `starts`, `ends`, `domain`, `ip`, `exists`). Matcher with `exists` and `ip` methods must be used without `value` and `ignore_case` fields.
- `name` (String) Name of the argument.

Optional Attributes:

- `ignore_case` (Boolean) Defines if case should be ignored. Defaults to `false`.
- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.
- `value` (String) String value.


<a id="nestedblock--matchers--response_header"></a>
### Nested Schema for `matchers.response_header`

Required Attributes:

- `method` (String) Match method (`exact`, `substring`, `regexp`, `starts`, `ends`, `domain`, `ip`, `exists`). Matcher with `exists` and `ip` methods must be used without `value` and `ignore_case` fields.
- `name` (String) Name of the argument.

Optional Attributes:

- `ignore_case` (Boolean) Defines if case should be ignored. Defaults to `false`.
- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.
- `value` (String) String value.


<a id="nestedblock--matchers--src_ip"></a>
### Nested Schema for `matchers.src_ip`

Required Attributes:

- `value` (String) IP address. CIDR masks are supported, e.g. `192.168.0.0/24`.

Optional Attributes:

- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.


<a id="nestedblock--matchers--src_port"></a>
### Nested Schema for `matchers.src_port`

Required Attributes:

- `method` (String) Match method (`equal`, `greater`, `greater_or_equal`, `less`, `less_or_equal`).
- `value` (Number) Integer value.

Optional Attributes:

- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.


<a id="nestedblock--matchers--src_port_range"></a>
### Nested Schema for `matchers.src_port_range`

Required Attributes:

- `range_end` (Number) Integer value.
- `range_start` (Number) Integer value.

Optional Attributes:

- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.


<a id="nestedblock--matchers--url"></a>
### Nested Schema for `matchers.url`

Required Attributes:

- `method` (String) Match method (`exact`, `substring`, `regexp`, `starts`, `ends`, `domain`, `ip`, `exists`). Matcher with `exists` and `ip` methods must be used without `value` and `ignore_case` fields.

Optional Attributes:

- `ignore_case` (Boolean) Defines if case should be ignored. Defaults to `false`.
- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.
- `value` (String) String value.


<a id="nestedblock--matchers--url_param"></a>
### Nested Schema for `matchers.url_param`

Required Attributes:

- `method` (String) Match method (`exact`, `substring`, `regexp`, `starts`, `ends`, `domain`, `ip`, `exists`). Matcher with `exists` and `ip` methods must be used without `value` and `ignore_case` fields.
- `name` (String) Name of the argument.

Optional Attributes:

- `ignore_case` (Boolean) Defines if case should be ignored. Defaults to `false`.
- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.
- `value` (String) String value.


<a id="nestedblock--matchers--url_query"></a>
### Nested Schema for `matchers.url_query`

Required Attributes:

- `method` (String) Match method (`exact`, `substring`, `regexp`, `starts`, `ends`, `domain`, `ip`, `exists`). Matcher with `exists` and `ip` methods must be used without `value` and `ignore_case` fields.

Optional Attributes:

- `ignore_case` (Boolean) Defines if case should be ignored. Defaults to `false`.
- `inverse` (Boolean) Defines if the condition should be inverted. Works similarly to logical NOT operator.
- `value` (String) String value.
