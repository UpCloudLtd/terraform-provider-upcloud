resource "upcloud_network" "net" {
  name = "network"
  zone = "{{.zone}}"

  ip_network {
    address            = "10.0.0.0/24"
    dhcp               = true
    dhcp_default_route = false
    family             = "IPv4"
    gateway            = "10.0.0.1"
  }
}
