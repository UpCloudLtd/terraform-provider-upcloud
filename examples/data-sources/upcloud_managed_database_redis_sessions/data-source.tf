# Use data source to gather a list of the active sessions for a Managed Redis Database

# Create a Managed Redis resource
resource "upcloud_managed_database_redis" "example" {
  name  = "example"
  title = "example"
  plan  = "1x1xCPU-2GB"
  zone  = "fi-hel2"
}

# Read the active sessions of the newly created service
data "upcloud_managed_database_redis_sessions" "example" {
  service = upcloud_managed_database_redis.example.id
}
