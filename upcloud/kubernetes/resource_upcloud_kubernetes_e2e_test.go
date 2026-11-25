package kubernetestests

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func testEndToEndKubernetes(t *testing.T, cidr string, privateNodeGroups bool) {
	serviceType := "NodePort"
	nodeAccess := "public"
	if privateNodeGroups {
		serviceType = "LoadBalancer"
		nodeAccess = "private"
	}

	t.Logf(`This testcase:

- Creates a Kubernetes cluster with one %s node group.
- Configures Kubernetes provider to connect to the created cluster using ephemeral cluster resource.
- Deploys hello deployment and service to the cluster.
- Uses http data source to verify that the deployment is reachable through a %s.
`, nodeAccess, serviceType)

	testdata := utils.ReadTestDataFile(t, "testdata/kubernetes_e2e.tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "~> 3.4",
			},
			"kubernetes": {
				VersionConstraint: "~> 3.0",
			},
		},
		Steps: []resource.TestStep{
			{
				// Create the cluster first and add kubernetes resources in the next step.
				Config: testdata,
				ConfigVariables: map[string]config.Variable{
					"private_node_groups": config.BoolVariable(privateNodeGroups),
					"network_cidr":        config.StringVariable(cidr),
				},
			},
			{
				Config: testdata,
				ConfigVariables: map[string]config.Variable{
					"enable_kubernetes_resources": config.BoolVariable(true),
					"network_cidr":                config.StringVariable(cidr),
					"private_node_groups":         config.BoolVariable(privateNodeGroups),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("upcloud_kubernetes_cluster.main", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("upcloud_kubernetes_node_group.default", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.http.hello.0", "status_code", "200"),
				),
			},
		},
	})
}

func TestEndToEndKubernetes_PublicNodePort(t *testing.T) {
	testEndToEndKubernetes(t, "172.23.45.0/24", false)
}

func TestEndToEndKubernetes_PrivateLoadBalancer(t *testing.T) {
	testEndToEndKubernetes(t, "172.23.46.0/24", true)
}

func TestEndToEndKubernetes_Authentication(t *testing.T) {
	t.Log(`This testcase:

1. Creates a Kubernetes cluster with a node group.
2. Deploys Keycloak to the cluster and exposes it with a load balancer.
3. Configures a realm, client and user to Keycloak and requests an ID token for the user.
4. Configures OIDC authentication for the cluster using the Keycloak as the issuer.
5. Uses the ID token to deploy a resource to the cluster as the OIDC user.
`)

	step := func(n int, check ...resource.TestCheckFunc) resource.TestStep {
		return resource.TestStep{
			// The configuration directory has provider blocks, so the provider factories need to be defined at the step level instead of the test case level.
			ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
			ConfigDirectory:          config.StaticDirectory("testdata/kubernetes_auth_e2e"),
			ConfigVariables: map[string]config.Variable{
				"step": config.IntegerVariable(n),
			},
			Check: resource.ComposeTestCheckFunc(check...),
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { upcloud.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			step(1),
			step(2),
			step(3),
			// Wait for the authentication configuration to take effect in the cluster before using it in the next step.
			step(4),
			step(5, resource.TestCheckResourceAttr("kubernetes_deployment_v1.hello.0", "metadata.0.name", "hello")),
		},
	})
}
