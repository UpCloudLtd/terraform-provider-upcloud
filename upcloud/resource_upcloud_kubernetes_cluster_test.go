// package upcloud

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"testing"

// 	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
// 	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
// 	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
// )

// func TestAccUpcloudKubernetesCluster(t *testing.T) {
// 	testClusterDataS1, err := os.ReadFile("testdata/upcloud_kubernetes/kubernetes_s1.tf")
// 	if err != nil {
// 		t.Fatal(fmt.Errorf("could not load testdata: %w", err))
// 	}

// 	resourceName := "upcloud_kubernetes_cluster.main"
// 	dataSourceName := "data.upcloud_kubernetes_cluster.main"

// 	var providers []*schema.Provider

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:          func() { testAccPreCheck(t) },
// 		ProviderFactories: testAccProviderFactories(&providers),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: string(testClusterDataS1),
// 				Check: resource.ComposeTestCheckFunc(
// 					// Basic checks via API
// 					testAccCheckUpcloudKubernetesClusterExistsViaAPI(resourceName),
// 					testAccCheckUpcloudKubernetesClusterIsReadyViaAPI(resourceName),

// 					// Check the state of the cluster resource
// 					resource.TestCheckResourceAttr(resourceName, "name", "tf-acc-test"),
// 					resource.TestCheckResourceAttrSet(resourceName, "network"),
// 					resource.TestCheckResourceAttrSet(resourceName, "network_cidr"),
// 					resource.TestCheckResourceAttr(resourceName, "state", "running"),
// 					resource.TestCheckResourceAttr(resourceName, "zone", "de-fra1"),

// 					// Check the state of cluster data source
// 					resource.TestCheckResourceAttrSet(dataSourceName, "client_certificate"),
// 					resource.TestCheckResourceAttrSet(dataSourceName, "client_key"),
// 					resource.TestCheckResourceAttrSet(dataSourceName, "cluster_ca_certificate"),
// 					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
// 					resource.TestCheckResourceAttrSet(dataSourceName, "host"),
// 					resource.TestCheckResourceAttrSet(dataSourceName, "kubeconfig"),

// 					// Check state for kubernetes plans data source
// 					resource.TestCheckResourceAttr("data.upcloud_kubernetes_plan.small", "description", "K8S-2xCPU-4GB"),
// 				),
// 			},
// 		},
// 	})
// }

// func testAccCheckUpcloudKubernetesClusterExistsViaAPI(name string) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		// retrieve the resource by name from state
// 		rs, ok := s.RootModule().Resources[name]
// 		if !ok {
// 			return fmt.Errorf("cluster not found in state when checking existence: %s", name)
// 		}

// 		if rs.Primary.ID == "" {
// 			return fmt.Errorf("cluster id is not set: %s", name)
// 		}

// 		svc, ok := testAccProvider.Meta().(*service.ServiceContext)
// 		if !ok {
// 			return fmt.Errorf("upcloud service unavailable when checking for cluster existence")
// 		}

// 		resp, err := svc.GetKubernetesCluster(context.Background(), &request.GetKubernetesClusterRequest{
// 			UUID: rs.Primary.ID,
// 		})
// 		if err != nil {
// 			return fmt.Errorf("getting cluster failed: %w", err)
// 		}

// 		if resp.UUID != rs.Primary.ID {
// 			return fmt.Errorf("cluster id does not match: expected: %s, actual: %s", rs.Primary.ID, resp.UUID)
// 		}

// 		return nil
// 	}
// }

// func testAccCheckUpcloudKubernetesClusterIsReadyViaAPI(name string) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[name]
// 		if !ok {
// 			return fmt.Errorf("cluster not found in state when checking readiness: %s", name)
// 		}

// 		svc, ok := testAccProvider.Meta().(*service.ServiceContext)
// 		if !ok {
// 			return fmt.Errorf("upcloud service unavailable when checking readiness")
// 		}

// 		resp, err := svc.GetKubernetesCluster(context.Background(), &request.GetKubernetesClusterRequest{
// 			UUID: rs.Primary.ID,
// 		})
// 		if err != nil {
// 			return fmt.Errorf("Error while fetching cluster details (%s)", name)
// 		}

// 		if resp.State != upcloud.KubernetesClusterStateRunning {
// 			return fmt.Errorf("cluster is not ready: %s", resp.State)
// 		}

// 		return nil
// 	}
// }
