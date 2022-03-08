package upcloud

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"math/rand"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	AlpineURL          = "https://dl-cdn.alpinelinux.org/alpine/v3.12/releases/x86/alpine-standard-3.12.0-x86.iso"
	AlpineHash         = "fd805e748f1950a34e354dc8fdfdf2f883237d65f5cdb8bcb47c64b0561d97a5"
	StorageTier        = "maxiops"
	storageDescription = "My data collection"
)

func TestAccUpcloudStorage_basic(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "upcloud_storage" "my_storage" {
						size  = 10
						tier  = "maxiops"
						title = "My_data"
						zone  = "pl-waw1"
						filesystem_autoresize = false
						delete_autoresize_backup = false

						backup_rule {
							interval = "daily"
							time = "2200"
							retention = 2
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_storage.my_storage", "size"),
					resource.TestCheckResourceAttrSet("upcloud_storage.my_storage", "tier"),
					resource.TestCheckResourceAttrSet("upcloud_storage.my_storage", "title"),
					resource.TestCheckResourceAttrSet("upcloud_storage.my_storage", "zone"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "size", "10"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "tier", "maxiops"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "title", "My_data"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "filesystem_autoresize", "false",
					),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "delete_autoresize_backup", "false",
					),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.interval", "daily"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.time", "2200"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.retention", "2"),
				),
			},
			{
				Config: `
					resource "upcloud_storage" "my_storage" {
						size  = 15
						tier  = "maxiops"
						title = "My_data_updated"
						zone  = "pl-waw1"
						filesystem_autoresize = true
						delete_autoresize_backup = true

						backup_rule {
							interval = "monday"
							time = "2230"
							retention = 5
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "size", "15"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "title", "My_data_updated"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "filesystem_autoresize", "true",
					),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "delete_autoresize_backup", "true",
					),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.interval", "weekly"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.time", "2230"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.retention", "5"),
				),
			},
		},
	})
}

func TestAccUpCloudStorage_import(t *testing.T) {
	var providers []*schema.Provider
	var storageDetails upcloud.StorageDetails

	expectedSize := "10"
	expectedTier := StorageTier
	expectedTitle := storageDescription
	expectedZone := "fi-hel1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfig(expectedSize, expectedTier, expectedTitle, expectedZone, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageExists("upcloud_storage.my_storage", &storageDetails),
				),
			},
			{
				ResourceName:      "upcloud_storage.my_storage",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUpCloudStorage_StorageImport(t *testing.T) {
	var providers []*schema.Provider
	var storageDetails upcloud.StorageDetails

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfigWithStorageImport(
					"http_import",
					AlpineURL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageExists("upcloud_storage.my_storage", &storageDetails),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "import.#", "1"),
					resource.TestCheckResourceAttr("upcloud_storage.my_storage", "import.0.sha256sum", AlpineHash),
				),
			},
		},
	})
}

func TestAccUpCloudStorage_StorageImportDirect(t *testing.T) {
	if os.Getenv(resource.TestEnvVar) != "" {
		var providers []*schema.Provider
		var storageDetails upcloud.StorageDetails

		imagePath, sum, err := createTempImage()
		if err != nil {
			t.Logf("unable to create temp image: %v", err)
			t.FailNow()
		}
		sha256sum := hex.EncodeToString((*sum).Sum(nil))

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories(&providers),
			CheckDestroy:      testAccCheckStorageDestroy,
			Steps: []resource.TestStep{
				{
					Config: testUpcloudStorageInstanceConfigWithStorageImport(
						"direct_upload",
						imagePath),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckStorageExists("upcloud_storage.my_storage", &storageDetails),
						resource.TestCheckResourceAttr(
							"upcloud_storage.my_storage", "import.#", "1"),
						resource.TestCheckResourceAttr("upcloud_storage.my_storage", "import.0.sha256sum", sha256sum),
					),
				},
			},
		})
	}
}

func TestAccUpCloudStorage_StorageImportValidation(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfigWithStorageImport(
					"gobbledigook",
					"somewhere"),
				ExpectError: regexp.MustCompile(`'source' value incorrect`),
			},
		},
	})
}

func TestAccUpCloudStorage_CloneImportValidation(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testUpcloudStorageInstanceConfigWithImportAndClone(),
				ExpectError: regexp.MustCompile("conflicts with"),
			},
		},
	})
}

func TestAccUpCloudStorage_CloneStorage(t *testing.T) {
	var providers []*schema.Provider
	var storageDetailsPlain upcloud.StorageDetails
	var storageDetailsClone upcloud.StorageDetails

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfigWithClone(20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageExists("upcloud_storage.plain_storage", &storageDetailsPlain),
					testAccCheckStorageExists("upcloud_storage.cloned_storage", &storageDetailsClone),
					resource.TestCheckResourceAttr(
						"upcloud_storage.cloned_storage", "clone.#", "1"),
					testAccCheckClonedStorageSize("upcloud_storage.cloned_storage", 20, &storageDetailsClone),
				),
			},
		},
	})
}

func testAccCheckClonedStorageSize(resourceName string, expected int, storage *upcloud.StorageDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Use the API SDK to locate the remote resource.
		client := testAccProvider.Meta().(*service.Service)
		latest, err := client.GetStorageDetails(&request.GetStorageDetailsRequest{
			UUID: storage.UUID,
		})

		if err != nil {
			return err
		}

		if latest.Size != expected {
			return fmt.Errorf("clone storage size is not as expected: %d != %d", expected, latest.Size)
		}

		return nil
	}
}

func testAccCheckStorageExists(resourceName string, storage *upcloud.StorageDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Look for the full resource name and error if not found
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// The provider has not set the ID for the resource
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Storage ID is set")
		}

		// Use the API SDK to locate the remote resource.
		client := testAccProvider.Meta().(*service.Service)
		latest, err := client.GetStorageDetails(&request.GetStorageDetailsRequest{
			UUID: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		// Update the reference the remote located storage
		*storage = *latest

		return nil
	}
}

func testAccCheckStorageDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "upcloud_storage" {
			continue
		}

		client := testAccProvider.Meta().(*service.Service)
		storages, err := client.GetStorages(&request.GetStoragesRequest{})
		if err != nil {
			return fmt.Errorf("[WARN] Error listing storage when deleting upcloud storage (%s): %s", rs.Primary.ID, err)
		}

		for _, storage := range storages.Storages {
			if storage.UUID == rs.Primary.ID {
				return fmt.Errorf("[WARN] Tried deleting Storage (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testUpcloudStorageInstanceConfig(size, tier, title, zone string, autoresize, deleteAutoresizeBackup bool) string {
	return fmt.Sprintf(`
		resource "upcloud_storage" "my_storage" {
			size  = %s
			tier  = "%s"
			title = "%s"
			zone  = "%s"
			filesystem_autoresize = %t
			delete_autoresize_backup = %t
		}
`, size, tier, title, zone, autoresize, deleteAutoresizeBackup)
}

func testUpcloudStorageInstanceConfigWithStorageImport(source, sourceLocation string) string {
	return fmt.Sprintf(`
		resource "upcloud_storage" "my_storage" {
			size  = 10
			tier  = "maxiops"
			title = "My Import Data"
			zone  = "fi-hel1"

			import {
				source = "%s"
				source_location = "%s"
			}
		}
`, source, sourceLocation)
}

func testUpcloudStorageInstanceConfigWithImportAndClone() string {
	return `
		resource "upcloud_storage" "my_storage" {
			size  = 10
			tier  = "maxiops"
			title = "My Imported with hash data"
			zone  = "fi-hel1"

			import {
				source = "foo"
				source_location = "bar"
				source_hash = "boo"
			}

			clone {
				id = "far"
			}
		}
	`
}

func testUpcloudStorageInstanceConfigWithClone(clonedSize int) string {
	return fmt.Sprintf(`
		resource "upcloud_storage" "plain_storage" {
			size  = 10
			tier  = "maxiops"
			title = "Plain storage"
			zone  = "fi-hel1"
		}

		resource "upcloud_storage" "cloned_storage" {
			size  = %d
			tier  = "maxiops"
			title = "My clone storage"
			zone  = "fi-hel1"

			clone {
				id = upcloud_storage.plain_storage.id
			}
		}
	`, clonedSize)
}

func createTempImage() (string, *hash.Hash, error) {
	imagePath := path.Join(os.TempDir(), fmt.Sprintf("temp_image_%s.img", acctest.RandString(5)))
	f, err := os.Create(imagePath)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	sum := sha256.New()
	for i := 0; i < 1000; i++ {
		b := []byte{byte(rand.Int())}
		_, err := f.Write(b)
		if err != nil {
			return "", nil, nil
		}
		_, err = sum.Write(b)
		if err != nil {
			return "", nil, err
		}
	}

	return imagePath, &sum, nil
}
