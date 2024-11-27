package upcloud

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUpcloudKubernetes(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_kubernetes/kubernetes_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_kubernetes/kubernetes_s2.tf")

	cName := "upcloud_kubernetes_cluster.main"
	g1Name := "upcloud_kubernetes_node_group.g1"
	g2Name := "upcloud_kubernetes_node_group.g2"
	g3Name := "upcloud_kubernetes_node_group.g3"

	verifyImportStep := func(name string, ignore ...string) resource.TestStep {
		return resource.TestStep{
			Config:                  testDataS1,
			ResourceName:            name,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: ignore,
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(cName, "control_plane_ip_filter.*", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(cName, "name", "tf-acc-test-k8s-cluster"),
					resource.TestCheckResourceAttr(cName, "version", "1.29"),
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
					resource.TestCheckResourceAttr(g2Name, "labels.%", "2"),
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
					}),
				),
			},
			{
				// Refresh state to include node-groups in the state.
				Config: testDataS1,
			},
			verifyImportStep(cName, "state"),
			verifyImportStep(g1Name),
			verifyImportStep(g2Name),
			verifyImportStep(g3Name),
			{
				Config: testDataS2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cName, "control_plane_ip_filter.#", "0"),
					resource.TestCheckResourceAttr(cName, "version", "1.29"),
					resource.TestCheckResourceAttr(g1Name, "node_count", "1"),
					resource.TestCheckResourceAttr(g2Name, "node_count", "2"),
				),
			},
		},
	})
}

func TestAccUpcloudKubernetes_labels(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_kubernetes/kubernetes_labels_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_kubernetes/kubernetes_labels_s2.tf")
	testDataS3 := utils.ReadTestDataFile(t, "testdata/upcloud_kubernetes/kubernetes_labels_s3.tf")

	cluster := "upcloud_kubernetes_cluster.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
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
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_kubernetes/kubernetes_storage_encryption_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_kubernetes/kubernetes_storage_encryption_s2.tf")
	nodeGroup := "upcloud_kubernetes_node_group.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
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
