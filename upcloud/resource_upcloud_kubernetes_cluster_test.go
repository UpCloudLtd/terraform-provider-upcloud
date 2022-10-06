package upcloud

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccUpcloudKubernetesCluster(t *testing.T) {
	basic, err := os.ReadFile("testdata/resource_upcloud_kubernetes_cluster/basic.tf")
	if err != nil {
		t.Fatal(fmt.Errorf("could not load testdata: %w", err))
	}

	full, err := os.ReadFile("testdata/resource_upcloud_kubernetes_cluster/full.tf")
	if err != nil {
		t.Fatal(fmt.Errorf("could not load testdata: %w", err))
	}

	resourceName := "upcloud_kubernetes_cluster.cluster"

	var providers []*schema.Provider
	var c upcloud.KubernetesCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: string(basic),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUpcloudKubernetesClusterExistsViaAPI(providers, resourceName, &c),
					testAccCheckUpcloudKubernetesClusterIsReadyViaAPI(&c),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "network"),
					resource.TestCheckResourceAttrSet(resourceName, "network_cidr"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "zone"),
				),
				Destroy: true,
			},
			{
				Config: string(full),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUpcloudKubernetesClusterExistsViaAPI(providers, resourceName, &c),
					testAccCheckUpcloudKubernetesClusterIsReadyViaAPI(&c),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "network"),
					resource.TestCheckResourceAttrSet(resourceName, "network_cidr"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "zone"),
				),
				Destroy: true,
			},
		},
	})
}

func testAccCheckUpcloudKubernetesClusterExistsViaAPI(providers []*schema.Provider, name string, c *upcloud.KubernetesCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("cluster not found in state: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("cluster id is not set: %s", name)
		}

		svc, ok := testAccProvider.Meta().(*service.ServiceContext)
		if !ok {
			return fmt.Errorf("provider client type assertion failed: upcloud")
		}

		resp, err := svc.GetKubernetesCluster(context.Background(), &request.GetKubernetesClusterRequest{
			UUID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("getting cluster failed: %w", err)
		}

		if resp == nil {
			return fmt.Errorf("cluster not found: %s", rs.Primary.ID)
		}

		if resp.UUID != rs.Primary.ID {
			return fmt.Errorf("cluster id does not match: expected: %s, actual: %s", rs.Primary.ID, resp.UUID)
		}

		// store the response
		c = resp

		return nil
	}
}

func testAccCheckUpcloudKubernetesClusterIsReadyViaAPI(c *upcloud.KubernetesCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if c == nil {
			return fmt.Errorf("cluster is nil")
		}

		if c.State != upcloud.KubernetesClusterStateRunning {
			return fmt.Errorf("cluster is not ready: %s", c.State)
		}

		return nil
	}
}
