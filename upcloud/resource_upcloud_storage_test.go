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

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	AlpineURL  = "https://dl-cdn.alpinelinux.org/alpine/v3.12/releases/x86/alpine-standard-3.12.0-x86.iso"
	AlpineHash = "fd805e748f1950a34e354dc8fdfdf2f883237d65f5cdb8bcb47c64b0561d97a5"
)

func TestAccUpcloudStorage_basic(t *testing.T) {
	var providers []*schema.Provider

	expectedSize := "10"
	expectedTier := "maxiops"
	expectedTitle := "My data collection"
	expectedZone := "fi-hel1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfig(expectedSize, expectedTier, expectedTitle, expectedZone),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_storage.my_storage", "size"),
					resource.TestCheckResourceAttrSet("upcloud_storage.my_storage", "tier"),
					resource.TestCheckResourceAttrSet("upcloud_storage.my_storage", "title"),
					resource.TestCheckResourceAttrSet("upcloud_storage.my_storage", "zone"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "size", expectedSize),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "tier", expectedTier),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "title", expectedTitle),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "zone", expectedZone),
				),
			},
		},
	})
}

func TestAccUpcloudStorage_basic_update(t *testing.T) {
	var providers []*schema.Provider

	expectedSize := "10"
	expectedTier := "maxiops"
	expectedTitle := "My data collection"
	expectedZone := "fi-hel1"

	expectedUpdatedSize := "20"
	expectedUpdatedTier := "hdd"
	expectedUpdatedTitle := "My Updated data collection"
	expectedUpdatedZone := "fi-hel2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfig(expectedSize, expectedTier, expectedTitle, expectedZone),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "size", expectedSize),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "tier", expectedTier),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "title", expectedTitle),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "zone", expectedZone),
				),
			},
			{
				Config: testUpcloudStorageInstanceConfig(expectedUpdatedSize, expectedUpdatedTier, expectedUpdatedTitle, expectedUpdatedZone),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "size", expectedUpdatedSize),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "tier", expectedUpdatedTier),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "title", expectedUpdatedTitle),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "zone", expectedUpdatedZone),
				),
			},
		},
	})
}

func TestAccUpcloudStorage_basic_backupRule(t *testing.T) {
	var providers []*schema.Provider

	expectedInterval := "daily"
	expectedTime := "2200"
	expectedRetention := "365"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfigWithBackupRule(expectedInterval, expectedTime, expectedRetention),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_storage.my_storage", "backup_rule.0.interval"),
					resource.TestCheckResourceAttrSet("upcloud_storage.my_storage", "backup_rule.0.time"),
					resource.TestCheckResourceAttrSet("upcloud_storage.my_storage", "backup_rule.0.retention"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.interval", expectedInterval),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.time", expectedTime),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.retention", expectedRetention),
				),
			},
		},
	})
}

func TestAccUpcloudStorage_backupRule_update(t *testing.T) {
	var providers []*schema.Provider

	expectedInterval := "daily"
	expectedTime := "2200"
	expectedRetention := "365"

	expectedUpdatedInterval := "thu"
	expectedUpdatedTime := "1300"
	expectedUpdatedRetention := "730"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfigWithBackupRule(expectedInterval, expectedTime, expectedRetention),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.interval", expectedInterval),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.time", expectedTime),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.retention", expectedRetention),
				),
			},
			{
				Config: testUpcloudStorageInstanceConfigWithBackupRule(expectedUpdatedInterval, expectedUpdatedTime, expectedUpdatedRetention),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.interval", expectedUpdatedInterval),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.time", expectedUpdatedTime),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my_storage", "backup_rule.0.retention", expectedUpdatedRetention),
				),
			},
		},
	})
}

func TestAccUpCloudStorage_import(t *testing.T) {
	var providers []*schema.Provider
	var storageDetails upcloud.StorageDetails

	expectedSize := "10"
	expectedTier := "maxiops"
	expectedTitle := "My data collection"
	expectedZone := "fi-hel1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfig(expectedSize, expectedTier, expectedTitle, expectedZone),
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

	resource.Test(t, resource.TestCase{
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

		resource.Test(t, resource.TestCase{
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

	resource.Test(t, resource.TestCase{
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

func TestAccUpCloudStorage_StorageImportDirectHash(t *testing.T) {
	if os.Getenv(resource.TestEnvVar) != "" {

		var providers []*schema.Provider
		var storageDetails1 upcloud.StorageDetails
		var storageDetails2 upcloud.StorageDetails

		imagePath, sum, err := createTempImage()
		if err != nil {
			t.Fatalf("unable to create temp image: %v", err)
		}
		sha256sum := hex.EncodeToString((*sum).Sum(nil))

		imagePath2, sum2, err := createTempImage()
		if err != nil {
			t.Fatalf("unable to create temp image: %v", err)
		}
		sha256sum2 := hex.EncodeToString((*sum2).Sum(nil))

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories(&providers),
			CheckDestroy:      testAccCheckStorageDestroy,
			Steps: []resource.TestStep{
				{
					Config: testUpcloudStorageInstanceConfigWithStorageImportHash(
						"direct_upload",
						imagePath,
						sha256sum),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckStorageExists("upcloud_storage.my_storage", &storageDetails1),
						resource.TestCheckResourceAttr(
							"upcloud_storage.my_storage", "import.#", "1"),
						resource.TestCheckResourceAttr("upcloud_storage.my_storage", "import.0.sha256sum", sha256sum),
						resource.TestCheckResourceAttrPair("upcloud_storage.my_storage", "import.0.sha256sum", "upcloud_storage.my_storage", "import.0.source_hash"),
					),
				},
				{
					PreConfig: func() {

						err := os.Remove(imagePath)

						if err != nil {
							t.Fatal(err)
						}

						err = os.Rename(imagePath2, imagePath)

						if err != nil {
							t.Fatal(err)
						}
					},
					Config: testUpcloudStorageInstanceConfigWithStorageImportHash(
						"direct_upload",
						imagePath,
						sha256sum2),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckStorageExists("upcloud_storage.my_storage", &storageDetails2),
						resource.TestCheckResourceAttr(
							"upcloud_storage.my_storage", "import.#", "1"),
						resource.TestCheckResourceAttr("upcloud_storage.my_storage", "import.0.sha256sum", sha256sum2),
						resource.TestCheckResourceAttrPair("upcloud_storage.my_storage", "import.0.sha256sum", "upcloud_storage.my_storage", "import.0.source_hash"),
						testAccCheckStorageDetailsDiffer(&storageDetails1, &storageDetails2),
					),
					ExpectNonEmptyPlan: true,
				},
			},
		})
	}
}

func TestAccUpCloudStorage_CloneImportValidation(t *testing.T) {
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testUpcloudStorageInstanceConfigWithImportAndClone(),
				ExpectError: regexp.MustCompile("ConflictsWith"),
			},
		},
	})
}

func TestAccUpCloudStorage_CloneStorage(t *testing.T) {
	var providers []*schema.Provider
	var storageDetailsPlain upcloud.StorageDetails
	var storageDetailsClone upcloud.StorageDetails

	resource.Test(t, resource.TestCase{
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

func testAccCheckStorageDetailsDiffer(d1 *upcloud.StorageDetails, d2 *upcloud.StorageDetails) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if d1.UUID == d2.UUID {
			return fmt.Errorf("old storage UUID unexpectedly matches new storage UUID: %s == %s", d1.UUID, d2.UUID)
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

func testUpcloudStorageInstanceConfig(size, tier, title, zone string) string {
	return fmt.Sprintf(`
		resource "upcloud_storage" "my_storage" {
			size  = %s
			tier  = "%s"
			title = "%s"
			zone  = "%s"
		}
`, size, tier, title, zone)
}

func testUpcloudStorageInstanceConfigWithBackupRule(interval, time, retention string) string {
	return fmt.Sprintf(`
		resource "upcloud_storage" "my_storage" {
			size  = 10
			tier  = "maxiops"
			title = "My data collection"
			zone  = "fi-hel1"
			backup_rule {
				interval = "%s"
				time = "%s"
				retention = "%s"
			}
		}
`, interval, time, retention)
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

func testUpcloudStorageInstanceConfigWithStorageImportHash(source, sourceLocation, sourceHash string) string {
	return fmt.Sprintf(`
		resource "upcloud_storage" "my_storage" {
			size  = 10
			tier  = "maxiops"
			title = "My Imported with hash data"
			zone  = "fi-hel1"

			import {
				source = "%s"
				source_location = "%s"
				source_hash = "%s"
			}
		}
`, source, sourceLocation, sourceHash)
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
	for i := 0; i < 100000000; i++ {
		b := []byte{byte(rand.Int())}
		_, err := f.Write(b)
		if err != nil {
			return "", nil, nil
		}
		sum.Write(b)
	}

	return imagePath, &sum, nil
}
