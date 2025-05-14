variable "prefix" {
  default = "tf-acc-test-objsto2-"
  type    = string
}

variable "zone" {
  default = "fi-hel1"
  type    = string
}

variable "region" {
  default = "europe-1"
  type    = string
}

resource "upcloud_managed_object_storage" "this" {
  name              = "tf-acc-test-objstov2-errors"
  region            = var.region
  configured_status = "started"

  labels = {
    // This is replaced during tests to test labels validation
    TEST_KEY = "TEST_VALUE"
  }
}
