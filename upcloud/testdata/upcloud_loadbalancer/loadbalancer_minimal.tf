variable "basename" {
  type    = string
  default = "tf-acc-test-lb-minimal"
}

variable "zone" {
  type    = string
  default = "fi-hel2"
}

resource "upcloud_loadbalancer" "this" {
  name              = "${var.basename}lb"
  plan              = "development"
  zone              = var.zone

  networks {
    type = "public"
    name = "public"
    family = "IPv4"
  }
}
