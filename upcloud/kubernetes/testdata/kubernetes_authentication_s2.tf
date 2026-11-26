variable "basename" {
  default = "tf-acc-test-k8s-auth-"
  type    = string
}

variable "zone" {
  default = "pl-waw1"
  type    = string
}

resource "upcloud_router" "main" {
  name = "${var.basename}router"
}

resource "upcloud_network" "main" {
  name   = "${var.basename}net"
  zone   = var.zone
  router = upcloud_router.main.id

  ip_network {
    address = "172.23.9.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_kubernetes_cluster" "main" {
  control_plane_ip_filter = ["0.0.0.0/0"]
  name                    = "${var.basename}cluster"
  network                 = upcloud_network.main.id
  zone                    = var.zone
}

resource "upcloud_kubernetes_cluster_authentication" "main" {
  cluster               = upcloud_kubernetes_cluster.main.id
  name                  = "example"
  issuer_url            = "https://accounts.example.com"
  audiences             = ["client-id-123", "client-id-456"]
  audience_match_policy = "MatchAny"

  claim_mappings = {
    username = {
      claim  = "email"
      prefix = "oidc:"
    }
    groups = {
      claim  = "groups"
      prefix = ""
    }
    extra = [
      {
        key              = "example.org/department"
        value_expression = "claims.department"
      }
    ]
  }

  claim_validation_rules = [
    {
      claim          = "hd"
      required_value = "example.com"
      message        = ""
    },
    {
      expression = "claims.email_verified == true"
      message    = "Email must be verified."
    }
  ]

  user_validation_rules = [
    {
      expression = "!user.username.startsWith('system:')"
      message    = "Username must not start with the system: prefix."
    }
  ]
}
