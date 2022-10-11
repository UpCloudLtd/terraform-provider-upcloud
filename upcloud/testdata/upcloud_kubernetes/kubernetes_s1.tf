variable "name" {
  default = "tf-acc-test"
  type    = string
}

variable "zone" {
  default = "de-fra1"
  type    = string
}

data "upcloud_kubernetes_plan" "small" {
  name = "small"
}

resource "upcloud_network" "main" {
  name = "terraform-provider-upcloud-test"
  zone = var.zone
  ip_network {
    address = "10.99.0.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_kubernetes_cluster" "main" {
  name    = var.name
  network = resource.upcloud_network.main.id
  zone    = var.zone

  node_group {
    count = 1
    labels = {
      env       = "dev"
      managedBy = "tf"
    }
    name = "small"
    plan = data.upcloud_kubernetes_plan.small.description
    ssh_keys = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIO3fnjc8UrsYDNU8365mL3lnOPQJg18V42Lt8U/8Sm+r testt_test"]
  }

  node_group {
    count = 1
    labels = {
      env       = "qa"
      managedBy = "tf"
    }
    name = "medium"
    plan = data.upcloud_kubernetes_plan.small.description
    ssh_keys = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIO3fnjc8UrsYDNU8365mL3lnOPQJg18V42Lt8U/8Sm+r testt_test"]
  }
}

data "upcloud_kubernetes_cluster" "main" {
  id = resource.upcloud_kubernetes_cluster.main.id
}
