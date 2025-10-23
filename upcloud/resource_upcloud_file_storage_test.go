package upcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// -----------------------------------------------------------------------------
// Basic Lifecycle Test
// -----------------------------------------------------------------------------
func TestAccUpCloudFileStorage_basicLifecycle(t *testing.T) {
	rName := fmt.Sprintf("file-storage-%s", acctest.RandString(5))
	netName := fmt.Sprintf("file-storage-net-%s", acctest.RandString(5))

	// Randomize subnet to avoid CIDR overlap conflicts
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("172.16.%d.0/24", subnet)
	storageIP := fmt.Sprintf("172.16.%d.11", subnet)
	aclTarget := fmt.Sprintf("172.16.%d.12", subnet)
	aclTarget2 := fmt.Sprintf("172.16.%d.15", subnet)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// -----------------------------------------------------------------
			// Step 1: Create
			// -----------------------------------------------------------------
			{
				Config: testAccFileStorageConfigStep1(rName, netName, cidr, storageIP, aclTarget),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccFileStorageExists("upcloud_file_storage.example"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "name", rName),
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
				Config: testAccFileStorageConfigStep2(rName, netName, cidr, storageIP, aclTarget, aclTarget2),
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
				Config: testAccFileStorageConfigStep3(netName, cidr, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccFileStorageExists("upcloud_file_storage.example"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "share.#", "0"),
					resource.TestCheckNoResourceAttr("upcloud_file_storage.example", "network"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "labels.single", "onlyone"),
					resource.TestCheckNoResourceAttr("upcloud_file_storage.example", "labels.environment"),
					resource.TestCheckNoResourceAttr("upcloud_file_storage.example", "labels.customer"),
				),
			},
			// -----------------------------------------------------------------
			// Step 4: Re-attach to network again (stabilize test cleanup)
			// -----------------------------------------------------------------
			{
				Config: testAccFileStorageConfigStep4(rName, netName, cidr, storageIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccFileStorageExists("upcloud_file_storage.example"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "network.family", "IPv4"),
					resource.TestCheckResourceAttr("upcloud_file_storage.example", "network.ip_address", storageIP),
				),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Import Test
// -----------------------------------------------------------------------------
func TestAccUpCloudFileStorage_import(t *testing.T) {
	rName := fmt.Sprintf("file-storage-import-%s", acctest.RandString(5))
	netName := fmt.Sprintf("file-storage-net-import-%s", acctest.RandString(5))

	// Randomize subnet to avoid conflicts
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("172.16.%d.0/24", subnet)
	storageIP := fmt.Sprintf("172.16.%d.11", subnet)
	aclTarget := fmt.Sprintf("172.16.%d.12", subnet)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccFileStorageConfigStep1(rName, netName, cidr, storageIP, aclTarget),
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

// -----------------------------------------------------------------------------
// Helper Config Builders
// -----------------------------------------------------------------------------

func testAccFileStorageConfigStep1(rName, netName, cidr, storageIP, aclTarget string) string {
	return fmt.Sprintf(`
		resource "upcloud_network" "this" {
			name = "%s"
			zone = "fi-hel2"

			ip_network {
				address = "%s"
				dhcp    = true
				family  = "IPv4"
			}
		}

		resource "upcloud_file_storage" "example" {
			name              = "%s"
			size              = 250
			zone              = "fi-hel2"
			configured_status = "stopped"

			labels = {
				environment = "staging"
				customer    = "example-customer"
			}

			share {
				name = "write-to-project"
				path = "/project"
				acl {
					target     = "%s"
					permission = "rw"
				}
			}

			network = {
				family     = "IPv4"
				name       = "example-private-net"
				uuid       = upcloud_network.this.id
				ip_address = "%s"
			}
		}
		`, netName, cidr, rName, aclTarget, storageIP)
}

func testAccFileStorageConfigStep2(rName, netName, cidr, storageIP, aclTarget, aclTarget2 string) string {
	return fmt.Sprintf(`
		resource "upcloud_network" "this" {
			name = "%s"
			zone = "fi-hel2"

			ip_network {
				address = "%s"
				dhcp    = true
				family  = "IPv4"
			}
		}

		resource "upcloud_file_storage" "example" {
			name              = "%s_v2"
			size              = 250
			zone              = "fi-hel2"
			configured_status = "started"

			labels = {
				environment = "staging"
				customer    = "example-customer"
				env         = "test"
			}

			share {
				name = "write-to-project"
				path = "/project"
				acl {
				target     = "%s"
				permission = "rw"
				}
			}

			share {
				name = "read-only"
				path = "/public"
				acl {
				target     = "%s"
				permission = "ro"
				}
			}

			network = {
				family     = "IPv4"
				name       = "example-private-net"
				uuid       = upcloud_network.this.id
				ip_address = "%s"
			}
		}
		`, netName, cidr, rName, aclTarget, aclTarget2, storageIP)
}

func testAccFileStorageConfigStep3(netName, cidr, rName string) string {
	return fmt.Sprintf(`
		resource "upcloud_network" "this" {
			name = "%s"
			zone = "fi-hel2"

			ip_network {
				address = "%s"
				dhcp    = true
				family  = "IPv4"
			}
		}
		resource "upcloud_file_storage" "example" {
			name              = "%s_v3"
			size              = 250
			zone              = "fi-hel2"
			configured_status = "started"

			labels = {
				single = "onlyone"
			}
		}
		`, netName, cidr, rName)
}

// Step 4 — Re-attach network to allow clean destroy
func testAccFileStorageConfigStep4(rName, netName, cidr, storageIP string) string {
	return fmt.Sprintf(`
		resource "upcloud_network" "this" {
			name = "%s"
			zone = "fi-hel2"

			ip_network {
				address = "%s"
				dhcp    = true
				family  = "IPv4"
			}
		}

		resource "upcloud_file_storage" "example" {
			name              = "%s_v4"
			size              = 250
			zone              = "fi-hel2"
			configured_status = "stopped"

			labels = {
				single = "onlyone"
			}

			network = {
				family     = "IPv4"
				name       = "example-private-net-readd"
				uuid       = upcloud_network.this.id
				ip_address = "%s"
			}
		}
	`, netName, cidr, rName, storageIP)
}

// -----------------------------------------------------------------------------
// Helper Check
// -----------------------------------------------------------------------------

func testAccFileStorageExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		return nil
	}
}
