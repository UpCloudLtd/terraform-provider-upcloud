variable "prefix" {
  default = "tf-acc-test-net-gateway-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

resource "upcloud_router" "this" {
  name = "${var.prefix}router"

  lifecycle {
    ignore_changes = [static_route]
  }
}

resource "upcloud_network" "this" {
  name = "${var.prefix}net"
  zone = var.zone

  ip_network {
    address = "172.16.124.0/24"
    dhcp    = true
    family  = "IPv4"
  }

  router = upcloud_router.this.id
}

resource "upcloud_gateway" "this" {
  name     = "${var.prefix}gw"
  zone     = var.zone
  plan     = "advanced"
  features = ["nat", "vpn"]

  router {
    id = upcloud_router.this.id
  }

  address {
    name = "my-public-ip"
  }

  labels = {
    test     = "net-gateway-tf"
    owned-by = "team-iaas"
  }
}

resource "upcloud_gateway_connection" "this" {
  gateway = upcloud_gateway.this.id
  name    = "test-connection"
  type    = "ipsec"

  local_route {
    name           = "local-route"
    type           = "static"
    static_network = "10.123.123.0/24"
  }

  remote_route {
    name           = "remote-route"
    type           = "static"
    static_network = "100.123.123.0/24"
  }
}

resource "upcloud_gateway_connection_tunnel" "this" {
  connection_id = upcloud_gateway_connection.this.id
  name       = "test-tunnel"
  local_address_name = tolist(upcloud_gateway.this.address).0.name
  remote_address = "100.123.123.10"
  
  ipsec_auth_psk {
    psk = "presharedkey1"
  }
}

resource "upcloud_gateway_connection_tunnel" "this2" {
  connection_id = upcloud_gateway_connection.this.id
  name       = "test-tunnel2"
  local_address_name = tolist(upcloud_gateway.this.address).0.name
  remote_address = "222.123.123.10"
  
  ipsec_auth_psk {
    psk = "i_like_cookies"
  }
}

resource "upcloud_gateway_connection" "this2" {
  gateway = upcloud_gateway.this.id
  name    = "test-connection2"
  type    = "ipsec"

  local_route {
    name           = "local-route2"
    type           = "static"
    static_network = "22.123.123.0/24"
  }

  remote_route {
    name           = "remote-route2"
    type           = "static"
    static_network = "222.123.123.0/24"
  }
}

