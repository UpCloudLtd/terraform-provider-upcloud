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

// ------------------------------------------------------------------------------
// Basic Lifecycle Test
// -----------------------------------------------------------------------------
func TestAccUpCloudFileStorage_basicLifecycle(t *testing.T) {
	configStep1 := utils.ReadTestDataFile(t, "testdata/upcloud_file_storage/file_storage_s1.tf")
	configStep2 := utils.ReadTestDataFile(t, "testdata/upcloud_file_storage/file_storage_s2.tf")
	configStep3 := utils.ReadTestDataFile(t, "testdata/upcloud_file_storage/file_storage_s3.tf")
	configStep4 := utils.ReadTestDataFile(t, "testdata/upcloud_file_storage/file_storage_s4.tf")

	prefix := "tf-acc-test-file-storage-"
	suffix := acctest.RandString(4)
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("172.16.%d.0/24", subnet)
	storageIP := fmt.Sprintf("172.16.%d.11", subnet)
	aclTarget := fmt.Sprintf("172.16.%d.12", subnet)
	aclTarget2 := fmt.Sprintf("172.16.%d.15", subnet)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configStep1,
				ConfigVariables: map[string]config.Variable{
					"prefix":           config.StringVariable(prefix),
					"suffix":           config.StringVariable(suffix),
					"cidr":             config.StringVariable(cidr),
					"network-ip-addrs": config.StringVariable(storageIP),
					"acl-1-ip":         config.StringVariable(aclTarget),
					"acl-2-ip":         config.StringVariable(aclTarget2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccFileStorageExists("upcloud_file_storage.this"),
					testAccFileStorageShareExists("upcloud_file_storage_share.this"),
					testAccFileStorageShareACLExists("upcloud_file_storage_share_acl.this"),
				),
			},
			{
				Config: configStep2,
				ConfigVariables: map[string]config.Variable{
					"prefix":           config.StringVariable(prefix),
					"suffix":           config.StringVariable(suffix),
					"cidr":             config.StringVariable(cidr),
					"network-ip-addrs": config.StringVariable(storageIP),
					"acl-1-ip":         config.StringVariable(aclTarget),
					"acl-2-ip":         config.StringVariable(aclTarget2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccFileStorageExists("upcloud_file_storage.this"),
					testAccFileStorageShareExists("upcloud_file_storage_share.this"),
					testAccFileStorageShareACLExists("upcloud_file_storage_share_acl.this"),
					testAccFileStorageShareExists("upcloud_file_storage_share.this2"),
					testAccFileStorageShareACLExists("upcloud_file_storage_share_acl.this2"),
				),
			},
			{
				Config: configStep3,
				ConfigVariables: map[string]config.Variable{
					"prefix":           config.StringVariable(prefix),
					"suffix":           config.StringVariable(suffix),
					"cidr":             config.StringVariable(cidr),
					"network-ip-addrs": config.StringVariable(storageIP),
					"acl-1-ip":         config.StringVariable(aclTarget),
					"acl-2-ip":         config.StringVariable(aclTarget2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccFileStorageExists("upcloud_file_storage.this"),
				),
			},
			{
				Config: configStep4,
				ConfigVariables: map[string]config.Variable{
					"prefix":           config.StringVariable(prefix),
					"suffix":           config.StringVariable(suffix),
					"cidr":             config.StringVariable(cidr),
					"network-ip-addrs": config.StringVariable(storageIP),
					"acl-1-ip":         config.StringVariable(aclTarget),
					"acl-2-ip":         config.StringVariable(aclTarget2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccFileStorageExists("upcloud_file_storage.this"),
				),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Import Test
// -----------------------------------------------------------------------------
func TestAccUpCloudFileStorage_import(t *testing.T) {
	configStep1 := utils.ReadTestDataFile(t, "testdata/upcloud_file_storage/file_storage_s1.tf")

	prefix := "tf-acc-test-file-storage-"
	suffix := acctest.RandString(4)
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("172.16.%d.0/24", subnet)
	storageIP := fmt.Sprintf("172.16.%d.11", subnet)
	aclTarget := fmt.Sprintf("172.16.%d.12", subnet)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configStep1,
				ConfigVariables: map[string]config.Variable{
					"prefix":           config.StringVariable(prefix),
					"suffix":           config.StringVariable(suffix),
					"cidr":             config.StringVariable(cidr),
					"network-ip-addrs": config.StringVariable(storageIP),
					"acl-1-ip":         config.StringVariable(aclTarget),
				},
			},
			{
				ResourceName:      "upcloud_file_storage.this",
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

func testAccFileStorageShareExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		return nil
	}
}

func testAccFileStorageShareACLExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		return nil
	}
}
