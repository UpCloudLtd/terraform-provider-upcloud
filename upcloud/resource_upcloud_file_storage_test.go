package upcloud

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// -----------------------------------------------------------------------------
// Basic Lifecycle Test
// -----------------------------------------------------------------------------
func TestAccUpCloudFileStorage_basicLifecycle(t *testing.T) {
	configStep1 := utils.ReadTestDataFile(t, "testdata/upcloud_file_storage/file_storage_cfg1.tf")
	configStep2 := utils.ReadTestDataFile(t, "testdata/upcloud_file_storage/file_storage_cfg2.tf")
	configStep3 := utils.ReadTestDataFile(t, "testdata/upcloud_file_storage/file_storage_cfg3.tf")
	configStep4 := utils.ReadTestDataFile(t, "testdata/upcloud_file_storage/file_storage_cfg4.tf")

	prefix := "tf-acc-test-file-storage-"
	rName := fmt.Sprintf("file-storage-%s", acctest.RandString(5))
	netName := fmt.Sprintf("file-storage-net-%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("172.16.%d.0/24", subnet)
	storageIP := fmt.Sprintf("172.16.%d.11", subnet)
	aclTarget := fmt.Sprintf("172.16.%d.12", subnet)
	aclTarget2 := fmt.Sprintf("172.16.%d.15", subnet)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		Steps: []resource.TestStep{
			// -----------------------------------------------------------------
			// Step 1: Create
			// -----------------------------------------------------------------
			{
				Config:          configStep1,
				ConfigVariables: stepVars(prefix, netName, rName, cidr, storageIP, aclTarget, aclTarget2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccFileStorageExists("upcloud_file_storage.example"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "name", prefix+rName),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "zone", "fi-hel2"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "size", "250"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "configured_status", "stopped"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "labels.environment", "staging"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "labels.customer", "example-customer"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "share.#", "1"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "share.0.acl.#", "1"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "share.0.acl.0.permission", "rw"),
				),
			},
			// -----------------------------------------------------------------
			// Step 2: Update (add label + new share + start service)
			// -----------------------------------------------------------------
			{
				Config:          configStep2,
				ConfigVariables: stepVars(prefix, netName, rName, cidr, storageIP, aclTarget, aclTarget2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccFileStorageExists("upcloud_file_storage.example"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "configured_status", "started"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "labels.environment", "staging"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "labels.customer", "example-customer"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "labels.env", "test"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "share.#", "2"),
				),
			},
			// -----------------------------------------------------------------
			// Step 3: Remove network and shares, replace labels completely
			// -----------------------------------------------------------------
			{
				Config:          configStep3,
				ConfigVariables: stepVars(prefix, netName, rName, cidr, storageIP, aclTarget, aclTarget2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccFileStorageExists("upcloud_file_storage.example"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "share.#", "0"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "network.#", "0"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "labels.single", "onlyone"),
					resource.TestCheckNoResourceAttr("upcloud_file_storage.example", "labels.environment"),
					resource.TestCheckNoResourceAttr("upcloud_file_storage.example", "labels.customer"),
				),
			},
			// -----------------------------------------------------------------
			// Step 4: Re-attach to network again
			// -----------------------------------------------------------------
			{
				Config:          configStep4,
				ConfigVariables: stepVars(prefix, netName, rName, cidr, storageIP, aclTarget, aclTarget2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccFileStorageExists("upcloud_file_storage.example"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "network.0.family", "IPv4"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "network.0.ip_address", storageIP),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "network.#", "1"),
				),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Import Test
// -----------------------------------------------------------------------------
func TestAccUpCloudFileStorage_import(t *testing.T) {
	configStep1 := utils.ReadTestDataFile(t, "testdata/upcloud_file_storage/file_storage_cfg1.tf")

	prefix := "tf-acc-test-file-storage-"
	rName := fmt.Sprintf("file-storage-import-%s", acctest.RandString(5))
	netName := fmt.Sprintf("file-storage-net-import-%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("172.16.%d.0/24", subnet)
	storageIP := fmt.Sprintf("172.16.%d.11", subnet)
	aclTarget := fmt.Sprintf("172.16.%d.12", subnet)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config:          configStep1,
				ConfigVariables: stepVars(prefix, netName, rName, cidr, storageIP, aclTarget, "none"),
			},
			// Import
			{
				ResourceName:      "upcloud_file_storage.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFileStorageExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		return nil
	}
}

func stepVars(prefix, netName, rName, cidr, storageIP, aclTarget, aclTarget2 string) map[string]config.Variable {
	return map[string]config.Variable{
		"prefix":            config.StringVariable(prefix),
		"net-name":          config.StringVariable(netName),
		"file-storage-name": config.StringVariable(rName),
		"cidr":              config.StringVariable(cidr),
		"network-ip-addrs":  config.StringVariable(storageIP),
		"acl-1-ip":          config.StringVariable(aclTarget),
		"acl-2-ip":          config.StringVariable(aclTarget2),
	}
}
