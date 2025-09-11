variable "basename" {
  default = "tf-acc-test-lb-"
  type    = string
}

variable "zone" {
  type    = string
  default = "fi-hel2"
}

locals {
  addresses = ["10.0.7.0/24", "10.0.8.0/24"]
}

resource "upcloud_network" "lb_network" {
  count = length(local.addresses)

  name = "${var.basename}net-${count.index}"
  zone = var.zone

  ip_network {
    address = local.addresses[count.index]
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_floating_ip_address" "ip" {
  count = 3

  access         = "public"
  family         = "IPv4"
  release_policy = "keep"
  zone           = var.zone
}

resource "upcloud_loadbalancer" "lb" {
  configured_status = "started"
  name              = "${var.basename}lb"
  plan              = "production-small"
  zone              = var.zone
  maintenance_dow   = "monday"
  maintenance_time  = "00:01:01Z"

  // Remove 1 IP and attach 1 new floating IP address
  ip_addresses = [
    for ip in slice(upcloud_floating_ip_address.ip, 1, 3) :
    {
      address      = ip.ip_address
      network_name = "public"
    }
  ]

  networks {
    type   = "public"
    family = "IPv4"
    name   = "public"
  }

  // Attach second private network
  dynamic "networks" {
    for_each = upcloud_network.lb_network

    content {
      name    = "lan-${networks.key}"
      type    = "private"
      family  = "IPv4"
      network = networks.value.id
    }
  }
}

resource "upcloud_loadbalancer_frontend" "lb_fe_1" {
  loadbalancer         = resource.upcloud_loadbalancer.lb.id
  name                 = "lb-fe-1-test"
  mode                 = "http"
  port                 = 8080
  default_backend_name = resource.upcloud_loadbalancer_backend.lb_be_1.name

  properties {
    timeout_client         = 20
    inbound_proxy_protocol = true
  }

  # change: add network listener
  networks {
    name = "lan-0"
  }
}

resource "upcloud_loadbalancer_backend" "lb_be_1" {
  loadbalancer  = resource.upcloud_loadbalancer.lb.id
  resolver_name = resource.upcloud_loadbalancer_resolver.lb_dns_1.name
  name          = "lb-be-1-test-1"
  properties {
    timeout_server          = 20
    timeout_tunnel          = 4000
    health_check_type       = "http"
    outbound_proxy_protocol = "v2"
  }
}

resource "upcloud_loadbalancer_frontend_rule" "lb_fe_1_r1" {
  frontend           = resource.upcloud_loadbalancer_frontend.lb_fe_1.id
  name               = "lb-fe-1-r1-test"
  priority           = 10
  matching_condition = "or"

  matchers {
    src_port {
      method  = "equal"
      value   = 80
      inverse = true
    }
    src_port_range {
      range_start = 100
      range_end   = 1000
      inverse     = true
    }
    src_ip {
      value   = "192.168.0.0/24"
      inverse = true
    }
    body_size {
      method  = "equal"
      value   = 8000
      inverse = true
    }
    body_size_range {
      range_start = 1000
      range_end   = 1001
      inverse     = true
    }
    path {
      method      = "starts"
      value       = "/application"
      ignore_case = true
      inverse     = true
    }
    url {
      method      = "starts"
      value       = "/application"
      ignore_case = true
      inverse     = true
    }
    url_query {
      method  = "starts"
      value   = "type=app"
      inverse = true
    }
    host {
      value   = "example.com"
      inverse = true
    }
    http_method {
      value   = "PATCH"
      inverse = true
    }
    cookie {
      method  = "exact"
      name    = "x-session-id"
      value   = "123456"
      inverse = true
    }
    header {
      method      = "exact"
      name        = "status"
      value       = "active"
      ignore_case = true
      inverse     = true
    }
    url_param {
      method      = "exact"
      name        = "status"
      value       = "active"
      ignore_case = true
      inverse     = true
    }
    num_members_up {
      method       = "less"
      value        = 1
      backend_name = resource.upcloud_loadbalancer_backend.lb_be_1.name
      inverse      = true
    }
  }
  actions {
    use_backend {
      backend_name = resource.upcloud_loadbalancer_backend.lb_be_1.name
    }
    http_redirect {
      location = "/app"
    }
    http_redirect {
      scheme = "https"
    }
    http_return {
      content_type = "text/plain"
      payload      = base64encode("Resource not found!")
      status       = "404"
    }
    tcp_reject {}
    set_forwarded_headers {}
  }
}

resource "upcloud_loadbalancer_resolver" "lb_dns_1" {
  loadbalancer  = resource.upcloud_loadbalancer.lb.id
  name          = "lb-resolver-1-test-1"
  cache_invalid = 10
  cache_valid   = 100
  retries       = 5
  timeout       = 10
  timeout_retry = 10
  nameservers   = ["94.237.127.9:53", "94.237.40.9:53"]
}
