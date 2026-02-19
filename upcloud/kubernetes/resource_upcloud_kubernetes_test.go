package kubernetestests

import (
	"context"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"golang.org/x/mod/semver"
)

// getLatestVersions fetches the latest Kubernetes versions from the UpCloud API
// and returns the two most recent versions as strings.
func getLatestVersions(t *testing.T) (string, string) {
	t.Helper()

	client := utils.NewServiceWithCredentialsFromEnv(t)
	versions, err := client.GetKubernetesVersions(context.Background(), &request.GetKubernetesVersionsRequest{})
	if err != nil {
		return "", ""
	}

	var v []string
	for _, version := range versions {
		v = append(v, version.Id)
	}
	semver.Sort(v)

	n := len(v)
	if n < 2 {
		return v[0], v[0]
	}
	return v[n-2], v[n-1]
}

func TestAccUpcloudKubernetes(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "../testdata/upcloud_kubernetes/kubernetes_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "../testdata/upcloud_kubernetes/kubernetes_s2.tf")

	cName := "upcloud_kubernetes_cluster.main"
	g1Name := "upcloud_kubernetes_node_group.g1"
	g2Name := "upcloud_kubernetes_node_group.g2"
	g3Name := "upcloud_kubernetes_node_group.g3"
	g4Name := "upcloud_kubernetes_node_group.g4"

	s1Version, s2Version := getLatestVersions(t)

	variables := func(version string) map[string]config.Variable {
		return map[string]config.Variable{
			"ver": config.StringVariable(version),
		}
	}

	verifyImportStep := func(name string, ignore ...string) resource.TestStep {
		return resource.TestStep{
			Config:                  testDataS1,
			ConfigVariables:         variables(s1Version),
			ResourceName:            name,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: ignore,
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testDataS1,
				ConfigVariables: variables(s1Version),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(cName, "control_plane_ip_filter.*", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(cName, "name", "tf-acc-test-k8s-cluster"),
					resource.TestCheckResourceAttr(cName, "version", s1Version),
					resource.TestCheckResourceAttr(cName, "zone", "fi-hel2"),
					resource.TestCheckResourceAttr(g1Name, "name", "small"),
					resource.TestCheckResourceAttr(g2Name, "name", "medium"),
					resource.TestCheckResourceAttr(g1Name, "anti_affinity", "true"),
					resource.TestCheckResourceAttr(g2Name, "anti_affinity", "false"),
					resource.TestCheckResourceAttr(g1Name, "node_count", "2"),
					resource.TestCheckResourceAttr(g2Name, "node_count", "1"),
					resource.TestCheckResourceAttr(g1Name, "ssh_keys.#", "1"),
					resource.TestCheckResourceAttr(g2Name, "ssh_keys.#", "1"),
					resource.TestCheckResourceAttr(g1Name, "labels.%", "2"),
					resource.TestCheckResourceAttr(g2Name, "labels.%", "3"),
					resource.TestCheckResourceAttr(g1Name, "labels.env", "dev"),
					resource.TestCheckResourceAttr(g1Name, "labels.managedBy", "tf"),
					resource.TestCheckTypeSetElemAttr(g1Name, "ssh_keys.*", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIO3fnjc8UrsYDNU8365mL3lnOPQJg18V42Lt8U/8Sm+r testt_test"),
					resource.TestCheckTypeSetElemNestedAttrs(g1Name, "kubelet_args.*", map[string]string{
						"key":   "log-flush-frequency",
						"value": "5s",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(g1Name, "taint.*", map[string]string{
						"effect": "NoExecute",
						"key":    "taintKey",
						"value":  "taintValue",
					}),
					resource.TestCheckResourceAttr(g1Name, "utility_network_access", "true"),
					resource.TestCheckResourceAttr(g2Name, "utility_network_access", "false"),

					resource.TestCheckResourceAttr(g3Name, "name", "encrypted-custom"),
					resource.TestCheckResourceAttr(g3Name, "plan", "custom"),
					resource.TestCheckResourceAttr(g3Name, "storage_encryption", "data-at-rest"),
					resource.TestCheckResourceAttr(g3Name, "custom_plan.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(g3Name, "custom_plan.*", map[string]string{
						"cores":        "1",
						"memory":       "2048",
						"storage_size": "25",
						"storage_tier": "maxiops",
					}),
				),
			},
			{
				// Refresh state to include node-groups in the state.
				Config:          testDataS1,
				ConfigVariables: variables(s1Version),
			},
			verifyImportStep(cName, "state"),
			verifyImportStep(g1Name),
			verifyImportStep(g2Name),
			verifyImportStep(g3Name),
			{
				Config: testDataS2,
				// Test cluster upgrade
				ConfigVariables: variables(s2Version),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cName, "control_plane_ip_filter.#", "0"),
					resource.TestCheckResourceAttr(cName, "version", s2Version),
					resource.TestCheckResourceAttr(g1Name, "node_count", "1"),
					resource.TestCheckResourceAttr(g2Name, "node_count", "2"),

					// Cloud Native plan node group checks
					resource.TestCheckResourceAttr(g4Name, "name", "cn-50g-storage"),
					resource.TestCheckResourceAttr(g4Name, "plan", "CLOUDNATIVE-4xCPU-8GB"),
					resource.TestCheckResourceAttr(g4Name, "node_count", "1"),
					resource.TestCheckResourceAttr(g4Name, "cloud_native_plan.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(g4Name, "cloud_native_plan.*", map[string]string{
						"storage_size": "50",
						"storage_tier": "standard",
					}),
				),
			},
		},
	})
}

func TestAccUpcloudKubernetes_labels(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "../testdata/upcloud_kubernetes/kubernetes_labels_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "../testdata/upcloud_kubernetes/kubernetes_labels_s2.tf")
	testDataS3 := utils.ReadTestDataFile(t, "../testdata/upcloud_kubernetes/kubernetes_labels_s3.tf")

	cluster := "upcloud_kubernetes_cluster.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(cluster, "control_plane_ip_filter.*", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(cluster, "name", "tf-acc-test-k8s-labels-cluster"),
					resource.TestCheckResourceAttr(cluster, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(cluster, "labels.%", "1"),
					resource.TestCheckResourceAttr(cluster, "labels.test", "terraform-provider-acceptance-test"),
					resource.TestCheckResourceAttr(cluster, "storage_encryption", "data-at-rest"),
				),
			},
			{
				Config: testDataS2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cluster, "name", "tf-acc-test-k8s-labels-cluster"),
					resource.TestCheckResourceAttr(cluster, "labels.%", "2"),
					resource.TestCheckResourceAttr(cluster, "labels.test", "terraform-provider-acceptance-test"),
					resource.TestCheckResourceAttr(cluster, "labels.managed-by", "terraform"),
				),
			},
			{
				Config: testDataS3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cluster, "name", "tf-acc-test-k8s-labels-cluster"),
					resource.TestCheckResourceAttr(cluster, "labels.%", "0"),
				),
			},
		},
	})
}

func TestAccUpcloudKubernetes_storageEncryption(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "../testdata/upcloud_kubernetes/kubernetes_storage_encryption_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "../testdata/upcloud_kubernetes/kubernetes_storage_encryption_s2.tf")
	nodeGroup := "upcloud_kubernetes_node_group.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(nodeGroup, "storage_encryption", "data-at-rest"),
				),
			},
			{
				Config:            testDataS1,
				ResourceName:      nodeGroup,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testDataS2,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(nodeGroup, "storage_encryption", "data-at-rest"),
				),
			},
			{
				Config:            testDataS2,
				ResourceName:      nodeGroup,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestEndToEndKubernetes(t *testing.T) {
	t.Log(`This testcase:

- Creates a Kubernetes cluster with one node group.
- Configures Kubernetes provider to connect to the created cluster using ephemeral cluster resource.
- Deploys hello deployment and service to the cluster.
- Uses http data source to verify that the deployment is reachable through a node port.
`)

	testdata := utils.ReadTestDataFile(t, "../testdata/upcloud_kubernetes/kubernetes_e2e.tf")

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
				// OpenTofu adds open action for the ephemeral resource which causes the plan to be non-empty.
				ExpectNonEmptyPlan: upcloud.UsingOpenTofu(),
			},
			{
				Config: testdata,
				ConfigVariables: map[string]config.Variable{
					"enable_kubernetes_resources": config.BoolVariable(true),
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
				// OpenTofu adds open action for the ephemeral resource which causes the plan to be non-empty.
				ExpectNonEmptyPlan: upcloud.UsingOpenTofu(),
			},
		},
	})
}
