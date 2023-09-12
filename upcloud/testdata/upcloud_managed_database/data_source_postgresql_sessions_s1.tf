resource "upcloud_managed_database_postgresql" "postgresql_sessions" {
  name = "postgresql-sessions-test-1"
  plan = "1x1xCPU-2GB-25GB"
  zone = "fi-hel2"
}

data "upcloud_managed_database_postgresql_sessions" "postgresql_sessions" {
  service = upcloud_managed_database_postgresql.postgresql_sessions.id
}
