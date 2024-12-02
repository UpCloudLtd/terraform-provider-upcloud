variable "prefix" {
  default = "tf-acc-test-db-termination-protection-"
  type    = string
}

variable "url_prefix" {
  default = "termination-protection-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

variable "db_count" {
  default = 1
  type    = number
}

variable "termination_protection" {
  default = false
  type    = bool
}

variable "powered" {
  default = true
  type    = bool
}

resource "upcloud_managed_database_mysql" "this" {
  count = var.db_count

  name                   = "${var.url_prefix}pg-${count.index}"
  plan                   = "1x1xCPU-2GB-25GB"
  powered                = var.powered
  termination_protection = var.termination_protection
  title                  = "${var.prefix}pg-${count.index}"
  zone                   = var.zone
}
