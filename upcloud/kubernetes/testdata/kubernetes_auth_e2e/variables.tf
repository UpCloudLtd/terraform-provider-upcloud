variable "basename" {
  default = "tf-acc-test-"
  type    = string
}

variable "network_cidr" {
  default = "172.23.55.0/24"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

variable "oidc_username" {
  default = "e2etest"
  type    = string
}

variable "step" {
  type = number
  # 1. Create cluster.
  # 2. Deploy and expose keycloak.
  # 3. Create user for OIDC.
  # 4. Configure OIDC authentication for the cluster.
  # 5. Use OIDC kubeconfig to deploy a resource to the cluster.
}
