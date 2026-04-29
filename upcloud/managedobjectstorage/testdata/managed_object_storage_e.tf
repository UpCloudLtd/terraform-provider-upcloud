variable "prefix" {
  default = "tf-acc-test-objstov2-"
  type    = string
}

variable "zone" {
  default = "se-sto1"
  type    = string
}

variable "region" {
  default = "europe-3"
  type    = string
}

resource "upcloud_managed_object_storage" "this" {
  name              = "${var.prefix}errors"
  region            = var.region
  configured_status = "started"

  labels = {
    // This is replaced during tests to test labels validation
    TEST_KEY = "TEST_VALUE"
  }
}
