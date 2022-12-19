resource "upcloud_managed_database_postgresql" "pg1" {
  name                    = "tf-pg-test-1"
  plan                    = "1x1xCPU-2GB-25GB"
  title                   = "tf-test-updated-pg-1"
  zone                    = "pl-waw1"
  maintenance_window_time = "11:00:00"
  maintenance_window_dow  = "thursday"
  powered                 = false
  properties {
    ip_filter = []
    version   = 14
  }
}

resource "upcloud_managed_database_postgresql" "pg2" {
  name    = "tf-pg-test-2"
  plan    = "1x1xCPU-2GB-25GB"
  title   = "tf-test-updated-pg-2"
  zone    = "pl-waw1"
  powered = true
  properties {
    version = 14
  }
}

resource "upcloud_managed_database_mysql" "msql1" {
  name  = "tf-mysql-test-1"
  plan  = "1x1xCPU-2GB-25GB"
  title = "tf-test-updated-msql-1"
  zone  = "pl-waw1"
}

resource "upcloud_managed_database_logical_database" "logical_db_1" {
  service = upcloud_managed_database_mysql.msql1.id
  name    = "tf-test-updated-logical-db-1"
}

resource "upcloud_managed_database_redis" "r1" {
  name  = "tf-redis-test-1"
  plan  = "1x1xCPU-2GB"
  title = "tf-test-redis-title-1"
  zone  = "pl-waw1"
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
