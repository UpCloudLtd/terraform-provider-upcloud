variable "prefix" {
  default = "tf-acc-test-objsto2-custom-domain-"
  type    = string
}

variable "zone" {
  default = "fi-hel1"
  type    = string
}

variable "region" {
  default = "europe-2"
  type    = string
}

resource "upcloud_managed_object_storage" "this" {
  name              = "${var.prefix}-objsto"
  region            = var.region
  configured_status = "started"

  network {
    family = "IPv4"
    name   = "public"
    type   = "public"
  }
}

// Delete the custom domain
