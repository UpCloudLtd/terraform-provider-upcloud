variable "basename" {
  default = "tf-acc-test-lb-"
  type    = string
}

variable "zone" {
  type    = string
  default = "fi-hel2"
}

resource "upcloud_network" "lb_network" {
  name = "${var.basename}net"
  zone = var.zone
  ip_network {
    address = "10.0.7.0/24"
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

  networks {
    name = "public"
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
    outbound_proxy_protocol = ""
    sticky_session_cookie_name = ""
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
