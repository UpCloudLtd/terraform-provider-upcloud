variable "prefix" {
  default = "tf-acc-test-file-storage-"
  type    = string
}

variable "net-name" {
  default = "net-name"
  type    = string
}

variable "router-name" {
  default = "router-name"
  type    = string
}

variable "network-cidr" {
  default    = "10.0.0.1/24"
  type        = string
}

variable "gateway-ip" {
    default   = "10.0.0.1"
  type        = string
}

resource "upcloud_router" "r" {
  name = "${var.prefix}${var.router-name}"

  static_route {
    route   = "192.168.0.0/24"
    nexthop = "10.20.0.254"
  }
}

resource "upcloud_network" "test" {
  name   = "${var.prefix}${var.net-name}"
  zone   = "fi-hel1"
  router = upcloud_router.r.id

  ip_network {
    address            = var.network-cidr
    dhcp               = true
    dhcp_default_route = true
    family             = "IPv4"
    gateway            = var.gateway-ip

    dhcp_routes_configuration = {
      effective_routes_auto_population = {
        enabled = true
      }
    }
  }
}
