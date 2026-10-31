variable "prefix" {
  default = "tf-acc-test-postgresql-ip-filter-"
  type    = string
}

variable "zone" {
  default = "fi-hel1"
  type    = string
}

resource "upcloud_managed_database_postgresql" "ip_filter" {
  name  = "${var.prefix}db"
  title = "${var.prefix}db"
  plan  = "1x1xCPU-2GB-25GB"
  zone  = var.zone
  properties {
    ip_filter = ["10.0.0.1/32"]
  }
}

# A downstream resource keyed on the connection details, the shape that was
# rewritten on every in-place update before the plan modifiers.
resource "terraform_data" "endpoint" {
  input = {
    host = upcloud_managed_database_postgresql.ip_filter.service_host
    port = upcloud_managed_database_postgresql.ip_filter.service_port
  }
}
