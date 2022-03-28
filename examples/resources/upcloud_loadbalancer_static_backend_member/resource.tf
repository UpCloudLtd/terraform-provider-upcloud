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

resource "upcloud_loadbalancer_static_backend_member" "lb_be_1_sm_1" {
  backend      = resource.upcloud_loadbalancer_backend.lb_be_1.id
  name         = "lb-be-1-sm-1-test"
  ip           = "10.0.0.10"
  port         = 8000
  weight       = 0
  max_sessions = 0
  enabled      = true
}
