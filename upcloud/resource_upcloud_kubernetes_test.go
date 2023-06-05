package upcloud

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudKubernetes(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_kubernetes/kubernetes_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_kubernetes/kubernetes_s2.tf")

	var providers []*schema.Provider
	cName := "upcloud_kubernetes_cluster.main"
	g1Name := "upcloud_kubernetes_node_group.g1"
	g2Name := "upcloud_kubernetes_node_group.g2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cName, "name", "tf-acc-test"),
					resource.TestCheckResourceAttr(cName, "zone", "de-fra1"),
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
				),
			},
			{
				Config:            testDataS2,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(g1Name, "node_count", "1"),
					resource.TestCheckResourceAttr(g2Name, "node_count", "2"),
				),
			},
		},
	})
}
