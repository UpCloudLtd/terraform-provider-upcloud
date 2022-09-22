package upcloud

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDataSourceUpcloudKubernetesCluster_basic(t *testing.T) {
	basic, err := os.ReadFile("testdata/datasource_upcloud_kubernetes_cluster/basic.tf")
	if err != nil {
		t.Fatal(fmt.Errorf("could not load testdata: %w", err))
	}

	var providers []*schema.Provider

	resourceName := "data.upcloud_kubernetes_cluster.basic"
	expectedName := "acc-test-datasource-upcloud-kubernetes-cluster-basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: string(basic),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "client_certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "client_key"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "host"),
					resource.TestCheckResourceAttrSet(resourceName, "kubeconfig"),
					resource.TestCheckResourceAttr(resourceName, "name", expectedName),
				),
				Destroy: true,
			},
		},
	})
}
