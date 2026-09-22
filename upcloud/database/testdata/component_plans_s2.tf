variable "prefix" {
  type = string
}

resource "upcloud_managed_database_postgresql" "component_plan" {
  name             = "${var.prefix}-pg"
  title            = "${var.prefix}-pg-updated"
  zone             = "fi-hel1"
  plan_compute     = "rdb.standard.2CPU-8GB"
  plan_node_count  = 2
  plan_storage_gib = 140
  plan_backups     = "regular"
}

resource "upcloud_managed_database_mysql" "component_plan" {
  name             = "${var.prefix}-mysql"
  title            = "${var.prefix}-mysql-updated"
  zone             = "fi-hel1"
  plan_compute     = "rdb.standard.2CPU-8GB"
  plan_node_count  = 2
  plan_storage_gib = 140
  plan_backups     = "regular"
}
