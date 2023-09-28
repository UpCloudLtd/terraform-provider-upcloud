resource "upcloud_managed_database_redis" "redis_sessions" {
  name = "tf-acc-test-redis-sessions-1"
  plan = "1x1xCPU-2GB"
  zone = "fi-hel2"
}

data "upcloud_managed_database_redis_sessions" "redis_sessions" {
  service = upcloud_managed_database_redis.redis_sessions.id
}
