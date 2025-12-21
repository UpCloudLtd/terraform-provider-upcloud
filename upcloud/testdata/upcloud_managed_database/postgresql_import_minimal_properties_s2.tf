variable "prefix" {
  default = "tf-acc-test-postgresql-min-props-"
  type    = string
}

variable "zone" {
  default = "fi-hel1"
  type    = string
}

resource "upcloud_managed_database_postgresql" "props" {
  name  = "postgresql-props-test"
  title = "${var.prefix}db"
  plan  = "1x1xCPU-2GB-25GB"
  zone  = var.zone
  properties {
    version = 17
    service_log = false
  }
}
