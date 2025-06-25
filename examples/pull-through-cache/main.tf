terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "~> 3.7"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.1"
    }
    upcloud = {
      source  = "UpCloudLtd/upcloud"
      version = "~> 5.20"
    }
  }
}

provider "upcloud" {}

resource "random_password" "token" {
  length  = 32
  special = false
}

resource "tls_private_key" "key" {
  count = var.ssh_public_key != "" ? 0 : 1

  algorithm = "ED25519"
}

resource "upcloud_network" "this" {
  name = "${var.prefix}net"
  zone = var.zone

  ip_network {
    address = "172.24.1.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_server" "this" {
  hostname = "${var.prefix}vm"
  title    = "${var.prefix}vm"
  zone     = var.zone
  plan     = "2xCPU-4GB"
  metadata = true
  user_data = templatefile("${path.module}/user_data.sh", {
    token = random_password.token.result
  })

  login {
    keys = [
      var.ssh_public_key != "" ? file(var.ssh_public_key) : tls_private_key.key[0].public_key_openssh
    ]
  }

  network_interface {
    type    = "private"
    network = upcloud_network.this.id
  }

  network_interface {
    type = "public"
  }

  template {
    storage = "Debian GNU/Linux 12 (Bookworm)"
  }
}

resource "upcloud_loadbalancer" "this" {
  name = "${var.prefix}lb"
  zone = var.zone
  plan = "essentials"

  networks {
    name   = "public"
    family = "IPv4"
    type   = "public"
  }

  networks {
    name    = "private"
    family  = "IPv4"
    network = upcloud_network.this.id
    type    = "private"
  }
}

resource "upcloud_loadbalancer_backend" "ptc" {
  loadbalancer = upcloud_loadbalancer.this.id
  name         = "ptc"

  properties {
    timeout_server = 86400
  }
}

resource "upcloud_loadbalancer_static_backend_member" "ptc" {
  backend      = upcloud_loadbalancer_backend.ptc.id
  name         = "vm"
  ip           = upcloud_server.this.network_interface[0].ip_address
  port         = 80
  weight       = 100
  max_sessions = 1000
  enabled      = true
}

resource "upcloud_loadbalancer_frontend" "ptc" {
  loadbalancer         = upcloud_loadbalancer.this.id
  name                 = "ptc"
  default_backend_name = upcloud_loadbalancer_backend.ptc.name
  mode                 = "http"
  port                 = 443

  networks {
    name = "public"
  }

  properties {}
}

resource "upcloud_loadbalancer_dynamic_certificate_bundle" "ptc" {
  name = "${var.prefix}cert"
  hostnames = [
    upcloud_loadbalancer.this.networks[0].dns_name,
  ]
  key_type = "rsa"
}

resource "upcloud_loadbalancer_frontend_tls_config" "ptc" {
  frontend           = upcloud_loadbalancer_frontend.ptc.id
  name               = "ptc"
  certificate_bundle = upcloud_loadbalancer_dynamic_certificate_bundle.ptc.id
}
