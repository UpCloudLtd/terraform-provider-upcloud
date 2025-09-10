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
  names     = ["lan-a", "lan-b"]
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

  // Remove all IPs
  ip_addresses = []

  # change: network names
  dynamic "networks" {
    for_each = upcloud_network.lb_network

    content {
      name    = local.names[networks.key]
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
    name = "lan-a"
  }

  networks {
    name = "lan-b"
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
