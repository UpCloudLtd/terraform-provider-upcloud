resource "upcloud_managed_database_valkey" "valkey_sessions" {
  name  = "tf-acc-test-valkey-sessions-1"
  title = "tf-acc-test-valkey-sessions-1"
  plan  = "1x1xCPU-2GB"
  zone  = "fi-hel2"
}

data "upcloud_managed_database_valkey_sessions" "valkey_sessions" {
  service = upcloud_managed_database_valkey.valkey_sessions.id
}
