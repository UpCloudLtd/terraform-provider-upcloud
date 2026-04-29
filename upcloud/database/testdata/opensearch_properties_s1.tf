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

  properties {
    automatic_utility_network_ip_filter = false
    public_access                       = false
    version                             = "2.19"

    segrep {
      pressure_enabled          = true
      pressure_checkpoint_limit = 5
    }
  }
}

data "upcloud_managed_database_opensearch_indices" "opensearch_indices" {
  service = upcloud_managed_database_opensearch.opensearch_properties.id
}
