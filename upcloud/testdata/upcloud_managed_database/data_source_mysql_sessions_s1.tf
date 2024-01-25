resource "upcloud_managed_database_mysql" "mysql_sessions" {
  name  = "mysql-sessions-test-1"
  title = "mysql-sessions-test-1"
  plan  = "1x1xCPU-2GB-25GB"
  zone  = "fi-hel2"
}

data "upcloud_managed_database_mysql_sessions" "mysql_sessions" {
  service = upcloud_managed_database_mysql.mysql_sessions.id
}
