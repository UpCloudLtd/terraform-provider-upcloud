resource "upcloud_server" "test_backup_rule" {
  hostname = "tf-acc-test-backup-rule-removal"
  zone     = "pl-waw1"
  plan     = "1xCPU-1GB"
  metadata = true

  login {
    password_delivery = "none"
  }

  template {
    storage = "Ubuntu Server 24.04 LTS (Noble Numbat)"
    size    = 25

    backup_rule {
      interval  = "daily"
      time      = "0100"
      retention = 8
    }
  }

  network_interface {
    type = "public"
  }
}
