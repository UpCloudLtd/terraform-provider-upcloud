variable "prefix" {
  type = string
}

resource "upcloud_managed_database_postgresql" "component_plan" {
  name             = "${var.prefix}-pg"
  title            = "${var.prefix}-pg"
  zone             = "fi-hel1"
  plan_compute     = "rdb.standard.2CPU-8GB"
  plan_node_count  = 2
  plan_storage_gib = 120
  plan_backups     = "regular"
}

resource "upcloud_managed_database_mysql" "component_plan" {
  name             = "${var.prefix}-mysql"
  title            = "${var.prefix}-mysql"
  zone             = "fi-hel1"
  plan_compute     = "rdb.standard.2CPU-8GB"
  plan_node_count  = 2
  plan_storage_gib = 120
  plan_backups     = "regular"
}
