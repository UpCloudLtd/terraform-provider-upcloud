resource "upcloud_server" "server1" {
  hostname = "server"
  zone     = "{{.zone}}"
  cpu      = "1"
  mem      = "1024"
  template {
    storage                  = "Debian GNU/Linux 11 (Bullseye)"
    size                     = 10
    filesystem_autoresize    = true
    delete_autoresize_backup = true
    backup_rule {
      time      = "0100"
      interval  = "mon"
      retention = 2
    }
  }
  network_interface {
    type = "public"
  }
  network_interface {
    type = "utility"
  }
  network_interface {
    type    = "private"
    network = resource.upcloud_network.net.id
  }
}

resource "upcloud_server" "server2" {
  hostname = "server"
  zone     = "{{.zone}}"
  cpu      = "1"
  mem      = "1024"
  template {
    storage = "Debian GNU/Linux 11 (Bullseye)"
    size    = 10
  }
  network_interface {
    type = "public"
  }
  simple_backup {
    time = "2200"
    plan = "dailies"
  }
}
