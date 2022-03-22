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

resource "upcloud_loadbalancer_resolver" "lb_dns_1" {
  loadbalancer  = resource.upcloud_loadbalancer.lb.id
  name          = "lb-resolver-1-test"
  cache_invalid = 10
  cache_valid   = 100
  retries       = 5
  timeout       = 10
  timeout_retry = 10
  nameservers   = ["94.237.127.9:53", "94.237.40.9:53"]
}

resource "upcloud_loadbalancer_backend" "lb_be_1" {
  loadbalancer  = resource.upcloud_loadbalancer.lb.id
  resolver_name = resource.upcloud_loadbalancer_resolver.lb_dns_1.name
  name          = "lb-be-1-test"
}

resource "upcloud_loadbalancer_static_backend_member" "lb_be_1_sm_1" {
  loadbalancer = resource.upcloud_loadbalancer.lb.id
  backend_name = resource.upcloud_loadbalancer_backend.lb_be_1.name
  name         = "lb-be-1-sm-1-test"
  ip           = "10.0.0.10"
  port         = 8000
  weight       = 100
  max_sessions = 1000
  enabled      = true
}

resource "upcloud_loadbalancer_dynamic_backend_member" "lb_be_1_dm_1" {
  loadbalancer = resource.upcloud_loadbalancer.lb.id
  backend_name = resource.upcloud_loadbalancer_backend.lb_be_1.name
  name         = "lb-be-1-dm-1-test"
  weight       = 10
  max_sessions = 10
  enabled      = false
}

resource "upcloud_loadbalancer_backend" "lb_be_2" {
  loadbalancer = resource.upcloud_loadbalancer.lb.id
  name         = "lb-be-2-test"
}

resource "upcloud_loadbalancer_static_backend_member" "lb_be_2_sm_1" {
  loadbalancer = resource.upcloud_loadbalancer.lb.id
  backend_name = resource.upcloud_loadbalancer_backend.lb_be_2.name
  name         = "lb-be-2-sm-1-test"
  ip           = "10.0.0.10"
  port         = 8000
  weight       = 0
  max_sessions = 0
  enabled      = true
}
