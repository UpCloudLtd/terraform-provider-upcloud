# Get the cluster level kubeconfig
data "upcloud_kubernetes_kubeconfig" "cluster" {
  cluster = upcloud_kubernetes_cluster.example.id
}

# Get the kubeconfig for users authenticating through the OIDC issuer "example-issuer" of the cluster
data "upcloud_kubernetes_kubeconfig" "oidc" {
  cluster        = upcloud_kubernetes_cluster.example.id
  authentication = "example-issuer"
  audience       = "example-client-id"
}
