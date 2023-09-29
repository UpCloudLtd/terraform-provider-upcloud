resource "upcloud_managed_database_postgresql" "pg1" {
  name                    = "tf-pg-test-1"
  plan                    = "1x1xCPU-2GB-25GB"
  title                   = "tf-test-pg-1"
  zone                    = "pl-waw1"
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
  name    = "tf-pg-test-2"
  plan    = "1x1xCPU-2GB-25GB"
  title   = "tf-test-pg-2"
  zone    = "pl-waw1"
  powered = true
  properties {
    version = 14
  }
}

resource "upcloud_managed_database_mysql" "msql1" {
  name  = "tf-mysql-test-2"
  plan  = "1x1xCPU-2GB-25GB"
  title = "tf-test-msql-1"
  zone  = "pl-waw1"
}

resource "upcloud_managed_database_logical_database" "logical_db_1" {
  service = upcloud_managed_database_mysql.msql1.id
  name    = "tf-test-logical-db-1"
}

resource "upcloud_managed_database_redis" "r1" {
  name  = "tf-redis-test-1"
  plan  = "1x1xCPU-2GB"
  title = "tf-test-redis-1"
  zone  = "pl-waw1"
}

resource "upcloud_managed_database_opensearch" "o1" {
  name  = "tf-opensearch-test-1"
  plan  = "1x2xCPU-4GB-80GB-1D"
  title = "tf-test-opensearch-1"
  zone  = "pl-waw1"
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

resource "upcloud_managed_database_user" "db_user_3" {
  service  = upcloud_managed_database_redis.r1.id
  username = "somename"
  password = "Superpass123"
  redis_access_control {
    categories = ["+@set"]
    channels   = ["*"]
    commands   = ["+set"]
    keys       = ["key_*"]
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
