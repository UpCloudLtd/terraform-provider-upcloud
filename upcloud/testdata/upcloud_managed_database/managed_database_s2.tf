variable "prefix" {
  default = "tf-acc-test-db-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

resource "upcloud_router" "pg2" {
  name = "${var.prefix}router-pg2"
}

resource "upcloud_router" "r1" {
  name = "${var.prefix}router-r1"
}

resource "upcloud_router" "msql1" {
  name = "${var.prefix}router-msql1"
}

resource "upcloud_network" "pg2" {
  name = "${var.prefix}net-pg2"
  zone = var.zone

  ip_network {
    address = "172.18.101.0/24"
    dhcp    = false
    family  = "IPv4"
  }

  router = upcloud_router.pg2.id
}

resource "upcloud_network" "r1" {
  name = "${var.prefix}net-r1"
  zone = var.zone

  ip_network {
    address = "172.18.102.0/24"
    dhcp    = false
    family  = "IPv4"
  }

  router = upcloud_router.r1.id
}

resource "upcloud_network" "msql1" {
  name = "${var.prefix}net-msql1"
  zone = var.zone

  ip_network {
    address = "172.18.103.0/24"
    dhcp    = false
    family  = "IPv4"
  }

  router = upcloud_router.msql1.id
}

resource "upcloud_managed_database_postgresql" "pg1" {
  name                    = "${var.prefix}pg-1"
  plan                    = "1x1xCPU-2GB-25GB"
  title                   = "${var.prefix}pg-1-updated"
  zone                    = var.zone
  maintenance_window_time = "11:00:00"
  maintenance_window_dow  = "thursday"
  powered                 = false

  properties {
    ip_filter = []
    version   = 14
  }
}

resource "upcloud_managed_database_postgresql" "pg2" {
  name    = "${var.prefix}pg-2"
  plan    = "1x1xCPU-2GB-25GB"
  title   = "${var.prefix}pg-2-updated"
  zone    = var.zone
  powered = true

  properties {
    version = 14
  }

  // No change in network
  network {
    family = "IPv4"
    name   = "${var.prefix}net"
    type   = "private"
    uuid   = upcloud_network.pg2.id
  }
}

resource "upcloud_managed_database_mysql" "msql1" {
  name  = "${var.prefix}mysql-1"
  plan  = "1x1xCPU-2GB-25GB"
  title = "${var.prefix}mysql-1-updated"
  zone  = var.zone

  // Attach network in modify
  network {
    family = "IPv4"
    name   = "${var.prefix}net"
    type   = "private"
    uuid   = upcloud_network.msql1.id
  }
}

resource "upcloud_managed_database_logical_database" "logical_db_1" {
  service = upcloud_managed_database_mysql.msql1.id
  name    = "${var.prefix}logical-db-1-updated"
}

resource "upcloud_managed_database_redis" "r1" {
  name  = "${var.prefix}redis-1"
  plan  = "1x1xCPU-2GB"
  title = "${var.prefix}redis-1-updated"
  zone  = var.zone

  // No change in network
  network {
    family = "IPv4"
    name   = "${var.prefix}net"
    type   = "private"
    uuid   = upcloud_network.r1.id
  }

}

resource "upcloud_managed_database_user" "db_user_1" {
  service        = upcloud_managed_database_mysql.msql1.id
  username       = "somename"
  password       = "Superpass890"
  authentication = "caching_sha2_password"
}

resource "upcloud_managed_database_user" "db_user_2" {
  service  = upcloud_managed_database_postgresql.pg2.id
  username = "somename"
  password = "Superpass123"
  pg_access_control {
    allow_replication = true
  }
}

resource "upcloud_managed_database_user" "db_user_3" {
  service  = upcloud_managed_database_redis.r1.id
  username = "somename"
  password = "Superpass123"
  redis_access_control {
    keys = ["key*"]
  }
}
