resource "upcloud_server_group" "main" {
  title         = "main_group"
  anti_affinity = true
  labels = {
    "key1" = "val1"
    "key2" = "val2"
    "key3" = "val3"
  }
  members = [
    "00b51165-fb58-4b77-bb8c-552277be1764",
    "00d56575-3821-3301-9de4-2b2bc7e35pqf",
    "000012dc-fe8c-a3y6-91f9-0db1215c36cf"
  ]
}
