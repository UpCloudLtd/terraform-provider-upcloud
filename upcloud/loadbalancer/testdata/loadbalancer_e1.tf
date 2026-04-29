// Actions validation: Empty actions array

variable "basename" {
  default = "tf-acc-test-lb-errors"
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
    address = "10.0.9.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_loadbalancer" "lb" {
  configured_status = "started"
  name              = "${var.basename}lb"
  plan              = "development"
  zone              = var.zone

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
  name         = "lb-be-test"
}

resource "upcloud_loadbalancer_frontend" "lb_fe_1" {
  loadbalancer         = resource.upcloud_loadbalancer.lb.id
  name                 = "lb-fe-test"
  mode                 = "http"
  port                 = 8080
  default_backend_name = resource.upcloud_loadbalancer_backend.lb_be_1.name
}

resource "upcloud_loadbalancer_frontend_rule" "lb_fe_1_r1" {
  frontend = resource.upcloud_loadbalancer_frontend.lb_fe_1.id
  name     = "test_validate_http_redirect"
  priority = 10

  matchers {
    src_port {
      method = "equal"
      value  = 8080
    }
  }
  actions {
    # http_redirect {
    # }
  }
}
