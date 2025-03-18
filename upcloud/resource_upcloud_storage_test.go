package upcloud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"regexp"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	AlpineURL          = "https://dl-cdn.alpinelinux.org/alpine/v3.12/releases/x86/alpine-standard-3.12.0-x86.iso"
	AlpineHash         = "fd805e748f1950a34e354dc8fdfdf2f883237d65f5cdb8bcb47c64b0561d97a5"
	StorageTier        = "maxiops"
	storageDescription = "tf-acc-test-storage"
)

func TestAccUpCloudStorage_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "upcloud_storage" "this" {
						encrypt = true
						size    = 10
						tier    = "maxiops"
						title   = "tf-acc-test-storage-basic-with-a-title-consisting-of-64-characters-or-more"
						zone    = "pl-waw1"
						filesystem_autoresize = false
						delete_autoresize_backup = false

						backup_rule {
							interval = "daily"
							time = "2200"
							retention = 2
						}
					}

					resource "upcloud_storage" "this-standard" {
						encrypt = true
						size    = 10
						tier    = "standard"
						title   = "tf-acc-test-storage-standard"
						zone    = "pl-waw1"
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
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "encrypt", "true"),
					resource.TestCheckResourceAttrSet("upcloud_storage.this", "size"),
					resource.TestCheckResourceAttrSet("upcloud_storage.this", "tier"),
					resource.TestCheckResourceAttrSet("upcloud_storage.this", "title"),
					resource.TestCheckResourceAttrSet("upcloud_storage.this", "zone"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "size", "10"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "tier", "maxiops"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "title", "tf-acc-test-storage-basic-with-a-title-consisting-of-64-characters-or-more"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "filesystem_autoresize", "false",
					),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "delete_autoresize_backup", "false",
					),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "backup_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "backup_rule.0.interval", "daily"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "backup_rule.0.time", "2200"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "backup_rule.0.retention", "2"),

					resource.TestCheckResourceAttrSet("upcloud_storage.this-standard", "tier"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this-standard", "tier", "standard"),
				),
			},
			{
				Config: `
					resource "upcloud_storage" "this" {
						encrypt = true
						size    = 15
						tier    = "maxiops"
						title   = "tf-acc-test-storage-basic-updated"
						zone    = "pl-waw1"
						filesystem_autoresize = true
						delete_autoresize_backup = true

						backup_rule {
							interval = "mon"
							time = "2230"
							retention = 5
						}
					}

					resource "upcloud_storage" "this-standard" {
						encrypt = true
						size    = 10
						tier    = "standard"
						title   = "tf-acc-test-storage-standard"
						zone    = "pl-waw1"
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
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "encrypt", "true"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "size", "15"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "title", "tf-acc-test-storage-basic-updated"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "filesystem_autoresize", "true",
					),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "delete_autoresize_backup", "true",
					),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "backup_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "backup_rule.0.interval", "mon"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "backup_rule.0.time", "2230"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "backup_rule.0.retention", "5"),
				),
			},
		},
	})
}

func TestAccUpCloudStorage_import(t *testing.T) {
	var storageDetails upcloud.StorageDetails

	expectedSize := "10"
	expectedTier := StorageTier
	expectedTitle := storageDescription
	expectedZone := "fi-hel1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfig(expectedSize, expectedTier, expectedTitle, expectedZone, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageExists("upcloud_storage.this", &storageDetails),
				),
			},
			{
				ResourceName:      "upcloud_storage.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUpCloudStorage_ImportAndTemplatize(t *testing.T) {
	var storageDetails upcloud.StorageDetails

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfigWithStorageImport(
					"http_import",
					AlpineURL) + testUpcloudStorageTemplateConfig("upcloud_storage.this.id"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageExists("upcloud_storage.this", &storageDetails),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "import.#", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_storage_template.this", "type", "template"),
					resource.TestCheckResourceAttr("upcloud_storage.this", "import.0.sha256sum", AlpineHash),
				),
			},
		},
	})
}

func TestAccUpCloudStorage_StorageImportDirect(t *testing.T) {
	var storageDetails upcloud.StorageDetails

	imagePath, sum, err := createTempImage()
	if err != nil {
		t.Logf("unable to create temp image: %v", err)
		t.FailNow()
	}
	sha256sum := hex.EncodeToString((*sum).Sum(nil))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfigWithStorageImport(
					"direct_upload",
					imagePath,
					sha256sum+"  filename"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageExists("upcloud_storage.this", &storageDetails),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "import.#", "1"),
					resource.TestCheckResourceAttr("upcloud_storage.this", "import.0.sha256sum", sha256sum),
				),
			},
		},
	})
}

func TestAccUpCloudStorage_StorageImportDirectCompressed(t *testing.T) {
	// Do not prepare the testdata, if the test will be skipped by resource.ParallelTest
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf(
			"Acceptance tests skipped unless env '%s' set",
			resource.EnvTfAcc,
		)
		return
	}

	var storageDetails upcloud.StorageDetails

	imagePath, sum, err := createTempImage()
	if err != nil {
		t.Logf("unable to create temp image: %v", err)
		t.FailNow()
	}
	sha256sum := hex.EncodeToString((*sum).Sum(nil))

	err = exec.Command("xz", imagePath).Run()
	if err != nil {
		t.Logf("unable to compress temp image: %v", err)
		t.FailNow()
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfigWithStorageImport(
					"direct_upload",
					imagePath+".xz",
					sha256sum),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageExists("upcloud_storage.this", &storageDetails),
					resource.TestCheckResourceAttr(
						"upcloud_storage.this", "import.#", "1"),
					// The SHA256 sum should match the original file, not the compressed one.
					resource.TestCheckResourceAttr("upcloud_storage.this", "import.0.sha256sum", sha256sum),
				),
			},
		},
	})
}

func TestAccUpCloudStorage_StorageHashValidation(t *testing.T) {
	imagePath, _, err := createTempImage()
	if err != nil {
		t.Logf("unable to create temp image: %v", err)
		t.FailNow()
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfigWithStorageImport(
					"direct_upload",
					imagePath,
					"this-is-not-the-right-hash",
				),
				ExpectError: regexp.MustCompile("imported storage's SHA256 sum does not match the source_hash:"),
			},
		},
	})
}

func TestAccUpCloudStorage_StorageImportValidation(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfigWithStorageImport(
					"gobbledigook",
					"somewhere"),
				ExpectError: regexp.MustCompile(`value must be one of: \["direct_upload" "http_import"\], got: ".*"`),
			},
		},
	})
}

func TestAccUpCloudStorage_CloneImportValidation(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testUpcloudStorageInstanceConfigWithImportAndClone(),
				ExpectError: regexp.MustCompile(`Attribute ".*" cannot be specified when ".*" is specified`),
			},
		},
	})
}

func TestAccUpCloudStorage_CloneStorage(t *testing.T) {
	var storageDetailsPlain upcloud.StorageDetails
	var storageDetailsClone upcloud.StorageDetails

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfigWithClone(20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageExists("upcloud_storage.plain_storage", &storageDetailsPlain),
					testAccCheckStorageExists("upcloud_storage.cloned_storage", &storageDetailsClone),
					resource.TestCheckResourceAttr(
						"upcloud_storage.cloned_storage", "clone.#", "1"),
					testAccCheckClonedStorageSize(20, &storageDetailsClone),
				),
			},
		},
	})
}

func testAccCheckClonedStorageSize(expected int, storage *upcloud.StorageDetails) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		// Use the API SDK to locate the remote resource.
		client := testAccProvider.Meta().(*service.Service)
		latest, err := client.GetStorageDetails(context.Background(), &request.GetStorageDetailsRequest{
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
		latest, err := client.GetStorageDetails(context.Background(), &request.GetStorageDetailsRequest{
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
		storages, err := client.GetStorages(context.Background(), &request.GetStoragesRequest{})
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
		resource "upcloud_storage" "this" {
			size  = %s
			tier  = "%s"
			title = "%s"
			zone  = "%s"
			filesystem_autoresize = %t
			delete_autoresize_backup = %t
		}
`, size, tier, title, zone, autoresize, deleteAutoresizeBackup)
}

func testUpcloudStorageInstanceConfigWithStorageImport(source, sourceLocation string, sourceHash ...string) string {
	sourceHashRow := ""
	if len(sourceHash) > 0 {
		sourceHashRow = fmt.Sprintf(`source_hash = "%s"`, sourceHash[0])
	}

	return fmt.Sprintf(`
		resource "upcloud_storage" "this" {
			size  = 10
			tier  = "maxiops"
			title = "tf-acc-test-storage-import"
			zone  = "fi-hel1"

			import {
				source = "%s"
				source_location = "%s"
				%s
			}
		}
`, source, sourceLocation, sourceHashRow)
}

func testUpcloudStorageTemplateConfig(idReference string) string {
	return fmt.Sprintf(`
		resource "upcloud_storage_template" "this" {
			source_storage = %s
			title = "tf-acc-test-storage-template"
		}
`, idReference)
}

func testUpcloudStorageInstanceConfigWithImportAndClone() string {
	return `
		resource "upcloud_storage" "this" {
			size  = 10
			tier  = "maxiops"
			title = "tf-acc-test-storage-import-hash"
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
			title = "tf-acc-test-storage-plain"
			zone  = "fi-hel1"
		}

		resource "upcloud_storage" "cloned_storage" {
			size  = %d
			tier  = "maxiops"
			title = "tf-acc-test-storage-cloned"
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

// Test Storage Backup Resource
func TestAccUpCloudStorageBackup_basic(t *testing.T) {
	var storageDetails upcloud.StorageDetails
	var backupDetails upcloud.StorageDetails

	resourceName := "upcloud_storage_backup.this"
	storageName := "upcloud_storage.test"

	// Define test title updates
	initialBackupTitle := "tf-acc-test-storage-backup"
	updatedBackupTitle := "tf-acc-test-storage-backup-updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckStorageBackupDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create Storage and Backup
			{
				Config: testUpcloudStorageBackupConfig(initialBackupTitle),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStorageExists(storageName, &storageDetails),
					testAccCheckStorageExists(resourceName, &backupDetails),
					resource.TestCheckResourceAttr(resourceName, "title", initialBackupTitle),
				),
			},
			// Step 2: Update the Backup Title
			{
				Config: testUpcloudStorageBackupConfig(updatedBackupTitle),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", updatedBackupTitle),
				),
			},
		},
	})
}

// Ensure that storage backups are properly removed
func testAccCheckStorageBackupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "upcloud_storage_backup" {
			continue
		}

		client := testAccProvider.Meta().(*service.Service)
		_, err := client.GetStorageDetails(context.Background(), &request.GetStorageDetailsRequest{
			UUID: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("Backup still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

// Generate Terraform Configuration for Storage & Backup
func testUpcloudStorageBackupConfig(backupTitle string) string {
	return fmt.Sprintf(`
		// Step 1: Create a storage resource
		resource "upcloud_storage" "test" {
			size  = 10
			tier  = "maxiops"
			title = "tf-acc-test-storage"
			zone  = "fi-hel1"
		}

		// Step 2: Create a backup from the storage
		resource "upcloud_storage_backup" "this" {
			source_storage = upcloud_storage.test.id
			title          = "%s"
		}
	`, backupTitle)
}
