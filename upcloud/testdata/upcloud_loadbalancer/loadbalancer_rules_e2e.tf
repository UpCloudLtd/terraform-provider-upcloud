variable "basename" {
  default = "tf-acc-test-lb-rules-"
  type    = string
}

variable "zone" {
  type    = string
  default = "fi-hel2"
}

locals {
  header_and_header_body = "header-and-header (animal=cat AND color=blue)"
  header_or_header_body  = "header-or-header (animal=cat OR animal=dog)"
  default_body           = "default"
}

resource "upcloud_network" "this" {
  name = "${var.basename}net"
  zone = var.zone
  ip_network {
    address = "10.0.10.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_loadbalancer" "this" {
  name = "${var.basename}lb"
  plan = "development"
  zone = var.zone

  networks {
    type   = "public"
    family = "IPv4"
    name   = "public"
  }

  networks {
    type    = "private"
    family  = "IPv4"
    name    = "private"
    network = resource.upcloud_network.this.id
  }
}

resource "upcloud_loadbalancer_backend" "this" {
  loadbalancer = resource.upcloud_loadbalancer.this.id
  name         = "default"
}

resource "upcloud_loadbalancer_frontend" "this" {
  loadbalancer         = resource.upcloud_loadbalancer.this.id
  name                 = "default"
  mode                 = "http"
  port                 = 80
  default_backend_name = upcloud_loadbalancer_backend.this.name

  networks {
    name = "public"
  }
}

resource "upcloud_loadbalancer_frontend_rule" "header_and_header" {
  frontend = resource.upcloud_loadbalancer_frontend.this.id
  name     = "header-and-header"
  priority = 10

  matchers {
    header {
      method      = "exact"
      name        = "animal"
      value       = "cat"
      ignore_case = true
    }

    header {
      method      = "exact"
      name        = "color"
      value       = "blue"
      ignore_case = true
    }
  }

  actions {
    http_return {
      content_type = "text/plain"
      payload      = base64encode(local.header_and_header_body)
      status       = "200"
    }
  }
}

resource "upcloud_loadbalancer_frontend_rule" "header_or_header" {
  frontend           = resource.upcloud_loadbalancer_frontend.this.id
  name               = "header-or-header"
  priority           = 5
  matching_condition = "or"

  matchers {
    header {
      method      = "exact"
      name        = "animal"
      value       = "cat"
      ignore_case = true
    }

    header {
      method      = "exact"
      name        = "animal"
      value       = "dog"
      ignore_case = true
    }
  }

  actions {
    http_return {
      content_type = "text/plain"
      payload      = base64encode(local.header_or_header_body)
      status       = "200"
    }
  }
}

resource "upcloud_loadbalancer_frontend_rule" "default" {
  frontend = resource.upcloud_loadbalancer_frontend.this.id
  name     = "default"
  priority = 0

  actions {
    http_return {
      content_type = "text/plain"
      payload      = base64encode(local.default_body)
      status       = "404"
    }
  }
}

data "http" "default" {
  url = "http://${upcloud_loadbalancer.this.networks.0.dns_name}"

  request_headers = {
    "animal" = "cow"
    "color"  = "red"
  }

  depends_on = [upcloud_loadbalancer_frontend_rule.default]

  # Wait 10 minutes for the LB to be ready. Other http data sources depend on this one and will thus wait for this request to succeed.
  retry {
    attempts     = 60
    min_delay_ms = 10e3
  }

  lifecycle {
    postcondition {
      condition     = self.response_body == local.default_body
      error_message = "Unexpected response body."
    }
  }
}

data "http" "blue_cat" {
  url = data.http.default.url

  request_headers = {
    "animal" = "cat"
    "color"  = "blue"
  }

  lifecycle {
    postcondition {
      condition     = self.response_body == local.header_and_header_body
      error_message = "Unexpected response body."
    }
  }
}

data "http" "red_cat" {
  url = data.http.default.url

  request_headers = {
    "animal" = "cat"
    "color"  = "red"
  }

  lifecycle {
    postcondition {
      condition     = self.response_body == local.header_or_header_body
      error_message = "Unexpected response body."
    }
  }
}
