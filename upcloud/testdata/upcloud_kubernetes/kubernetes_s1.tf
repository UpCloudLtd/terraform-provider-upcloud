variable "basename" {
  default = "tf-acc-test-k8s-"
  type    = string
}

variable "zone" {
  default = "fi-hel2"
  type    = string
}

variable "ver" {
  default = "1.30"
  type    = string
}

resource "upcloud_network" "main" {
  name = "${var.basename}net"
  zone = var.zone
  ip_network {
    address = "172.23.7.0/24"
    dhcp    = true
    family  = "IPv4"
  }
  # UpCloud Kubernetes Service will add a router to this network to ensure cluster networking is working as intended.
  lifecycle {
    ignore_changes = [router]
  }
}

resource "upcloud_kubernetes_cluster" "main" {
  control_plane_ip_filter = ["0.0.0.0/0"]
  name                    = "${var.basename}cluster"
  network                 = upcloud_network.main.id
  version                 = var.ver
  zone                    = var.zone
}

resource "upcloud_kubernetes_node_group" "g1" {
  cluster       = upcloud_kubernetes_cluster.main.id
  node_count    = 2
  anti_affinity = true
  labels = {
    env       = "dev"
    managedBy = "tf"
  }
  name     = "small"
  plan     = "2xCPU-4GB"
  ssh_keys = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIO3fnjc8UrsYDNU8365mL3lnOPQJg18V42Lt8U/8Sm+r testt_test"]
  taint {
    effect = "NoExecute"
    key    = "taintKey"
    value  = "taintValue"
  }
  kubelet_args {
    key   = "log-flush-frequency"
    value = "5s"
  }
  utility_network_access = true
}

resource "upcloud_kubernetes_node_group" "g2" {
  cluster    = upcloud_kubernetes_cluster.main.id
  node_count = 1
  labels = {
    env       = "qa"
    managedBy = "tf"
  }
  name = "medium"
  plan = "2xCPU-4GB"
  ssh_keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIO3fnjc8UrsYDNU8365mL3lnOPQJg18V42Lt8U/8Sm+r testt_test"
  ]
  utility_network_access = false
}

resource "upcloud_kubernetes_node_group" "g3" {
  cluster    = upcloud_kubernetes_cluster.main.id
  node_count = 1
  name       = "encrypted-custom"
  plan       = "custom"
  custom_plan {
    cores        = 1
    memory       = 2048
    storage_size = 25
  }
  storage_encryption = "data-at-rest"
}

data "upcloud_kubernetes_cluster" "main" {
  id = upcloud_kubernetes_cluster.main.id
}
