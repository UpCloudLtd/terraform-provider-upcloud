# Use data source to gather a list of the active sessions for a Managed MySQL Database

# Create a Managed MySQL resource
resource "upcloud_managed_database_mysql" "example" {
  name  = "mysql-example1"
  title = "mysql-example1"
  plan  = "1x1xCPU-2GB-25GB"
  zone  = "fi-hel1"
}

# Read the active sessions of the newly created service
data "upcloud_managed_database_mysql_sessions" "example" {
  service = upcloud_managed_database_mysql.example.id
}
