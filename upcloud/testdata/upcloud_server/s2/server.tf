resource "upcloud_server" "server1" {
  hostname = "server"
  zone     = "{{.zone}}"
  cpu      = "2"
  mem      = "2048"
  template {
    storage                  = "Debian GNU/Linux 11 (Bullseye)"
    size                     = 20
    filesystem_autoresize    = true
    delete_autoresize_backup = true
    backup_rule {
      time      = "0200"
      interval  = "fri"
      retention = 1
    }
  }
  network_interface {
    type = "public"
  }
  network_interface {
    type = "utility"
  }
}
