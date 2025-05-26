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

resource "upcloud_router" "v1" {
  name = "${var.prefix}router-v1"
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

resource "upcloud_network" "v1" {
  name = "${var.prefix}net-v1"
  zone = var.zone

  ip_network {
    address = "172.18.104.0/24"
    dhcp    = false
    family  = "IPv4"
  }

  router = upcloud_router.v1.id
}

resource "upcloud_managed_database_postgresql" "pg1" {
  name                    = "${var.prefix}pg-1"
  plan                    = "1x1xCPU-2GB-25GB"
  title                   = "${var.prefix}pg-1"
  zone                    = var.zone
  maintenance_window_time = "10:00:00"
  maintenance_window_dow  = "friday"
  powered                 = true

  properties {
    public_access = true
    ip_filter     = ["10.0.0.1/32"]
    version       = 13
  }
}

resource "upcloud_managed_database_postgresql" "pg2" {
  name    = "${var.prefix}pg-2"
  plan    = "1x1xCPU-2GB-25GB"
  title   = "${var.prefix}pg-2"
  zone    = var.zone
  powered = true

  properties {
    version = 14
  }

  // Attach network on create
  network {
    family = "IPv4"
    name   = "${var.prefix}net-pg2"
    type   = "private"
    uuid   = upcloud_network.pg2.id
  }

  labels = {
    test = "terraform-provider-acceptance-test"
  }
}

resource "upcloud_managed_database_mysql" "msql1" {
  name  = "${var.prefix}mysql-1"
  plan  = "1x1xCPU-2GB-25GB"
  title = "${var.prefix}mysql-1"
  zone  = var.zone

  labels = {
    test       = ""
    managed-by = "team-devex"
  }
}

resource "upcloud_managed_database_logical_database" "logical_db_1" {
  service = upcloud_managed_database_mysql.msql1.id
  name    = "${var.prefix}logical-db-1"
}

resource "upcloud_managed_database_valkey" "v1" {
  name  = "${var.prefix}valkey-1"
  plan  = "1x1xCPU-2GB"
  title = "${var.prefix}valkey-1"
  zone  = var.zone

  // Attach network on create
  network {
    family = "IPv4"
    name   = "${var.prefix}net-v1"
    type   = "private"
    uuid   = upcloud_network.v1.id
  }
}

resource "upcloud_managed_database_opensearch" "o1" {
  name  = "${var.prefix}opensearch-1"
  plan  = "1x2xCPU-4GB-80GB-1D"
  title = "${var.prefix}opensearch-1"
  zone  = var.zone

  labels = {
    test = "terraform-provider-acceptance-test"
  }
}

resource "upcloud_managed_database_user" "db_user_1" {
  service        = upcloud_managed_database_mysql.msql1.id
  username       = "somename"
  password       = "Superpass123"
  authentication = "mysql_native_password"
}

resource "upcloud_managed_database_user" "db_user_2" {
  service  = upcloud_managed_database_postgresql.pg2.id
  username = "somename"
  password = "Superpass123"
  pg_access_control {
    allow_replication = false
  }
}

resource "upcloud_managed_database_user" "db_user_4" {
  service  = upcloud_managed_database_opensearch.o1.id
  username = "somename"
  password = "Superpass12345"
  opensearch_access_control {
    rules {
      index      = ".opensearch-observability"
      permission = "admin"
    }
  }
}

resource "upcloud_managed_database_user" "db_user_5" {
  service  = upcloud_managed_database_valkey.v1.id
  username = "somename"
  password = "Superpass123"
  valkey_access_control {
    categories = ["+@all"]
    channels   = ["*"]
    commands   = ["+set", "+get", "+del"]
    keys       = ["key_*"]
  }
}
