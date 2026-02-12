// This file is used for the all test steps in the termination protection test. The variables below to define the changes between test steps.

variable "basename" {
  default = "tf-acc-test-"
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

locals {
  name_prefix = "${var.basename}db-termination-protection-"
}

resource "upcloud_managed_database_mysql" "this" {
  count = var.db_count

  name                   = "${var.url_prefix}db-${count.index}"
  plan                   = "1x1xCPU-2GB-25GB"
  powered                = var.powered
  termination_protection = var.termination_protection
  title                  = "${local.name_prefix}${count.index}"
  zone                   = var.zone
}
