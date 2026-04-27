resource "upcloud_server_group" "this" {
  title         = "tf-acc-test-server-group-errors"
  track_members = false
  members       = ["621e1d72-6e83-43eb-ba95-37d98ade0fac"]
}
