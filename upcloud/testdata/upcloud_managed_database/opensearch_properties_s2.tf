variable "prefix" {
  default = "tf-acc-test-os-props-"
  type    = string
}

variable "zone" {
  default = "fi-hel2"
  type    = string
}

resource "upcloud_managed_database_opensearch" "opensearch_properties" {
  name  = "${var.prefix}db"
  title = "${var.prefix}db"
  plan  = "1x2xCPU-4GB-80GB-1D"
  zone  = var.zone

  access_control          = true
  extended_access_control = true
  properties {
    automatic_utility_network_ip_filter = true
    public_access                       = true

    segrep {
      pressure_checkpoint_limit = 6
    }
  }
}
