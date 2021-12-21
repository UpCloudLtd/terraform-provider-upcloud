# return all available networks
data "upcloud_networks" "upcloud" {}

# return all available networks within a zone
data "upcloud_networks" "upcloud_by_zone" {
  zone = "fi-hel1"
}

# return all available networks filtered by a regular expression on the name of the network
data "upcloud_networks" "upcloud_by_zone" {
  filter_name = "^Public.*"
}
