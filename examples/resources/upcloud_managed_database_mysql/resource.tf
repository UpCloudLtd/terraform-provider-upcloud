# Minimal config
resource "upcloud_managed_database_mysql" "example_1" {
  name  = "mysql-1"
  title = "mysql-1-example-1"
  plan  = "1x1xCPU-2GB-25GB"
  zone  = "fi-hel1"
}

# Shutdown instance after creation
resource "upcloud_managed_database_mysql" "example_2" {
  name    = "mysql-2"
  title   = "mysql-2-example-2"
  plan    = "1x1xCPU-2GB-25GB"
  zone    = "fi-hel1"
  powered = false
}

# Service with custom properties
# Note that this basically sets strict mode off which is not normally recommended
resource "upcloud_managed_database_mysql" "example_3" {
  name  = "mysql-3"
  title = "mysql-3-example-3"
  plan  = "1x1xCPU-2GB-25GB"
  zone  = "fi-hel1"
  properties {
    sql_mode           = "NO_ENGINE_SUBSTITUTION"
    wait_timeout       = 300
    sort_buffer_size   = 4e+6  # 4MB
    max_allowed_packet = 16e+6 # 16MB
    admin_username     = "admin"
    admin_password     = "<ADMIN_PASSWORD>"
  }
}
