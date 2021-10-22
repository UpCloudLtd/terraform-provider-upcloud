package upcloud

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccCheckStorageBackupRule(storageName string, backupRule *upcloud.BackupRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*service.Service)
		resourceName := fmt.Sprintf("upcloud_storage.%s", storageName)

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Storage with the name: %s is not set", storageName)
		}

		storageDetails, err := client.GetStorageDetails(&request.GetStorageDetailsRequest{UUID: rs.Primary.ID})
		if err != nil {
			return err
		}

		if storageDetails.BackupRule.Time != backupRule.Time || storageDetails.BackupRule.Retention != backupRule.Retention || storageDetails.BackupRule.Interval != backupRule.Interval {
			return fmt.Errorf("Storage backup rule does not mach. Exprected: %+v, received: %+v", storageDetails.BackupRule, backupRule)
		}

		return nil
	}
}

func TestAccUpCloudStorageBackup_basic(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "upcloud_storage" "s1" {
						title = "addon"
						size = 10
						zone = "pl-waw1"
					}

					resource "upcloud_storage_backup" "b1" {
						storage = upcloud_storage.s1.id
						time = "2200"
						interval = "mon"
						retention = 2
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_storage_backup.b1", "time", "2200"),
					resource.TestCheckResourceAttr("upcloud_storage_backup.b1", "interval", "mon"),
					resource.TestCheckResourceAttr("upcloud_storage_backup.b1", "retention", "2"),
					testAccCheckStorageBackupRule("s1", &upcloud.BackupRule{
						Time:      "2200",
						Interval:  "mon",
						Retention: 2,
					}),
				),
			},
			{
				Config: `
					resource "upcloud_storage" "s1" {
						title = "addon"
						size = 10
						zone = "pl-waw1"
					}

					resource "upcloud_storage_backup" "b1" {
						storage = upcloud_storage.s1.id
						time = "0000"
						interval = "daily"
						retention = 4
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_storage_backup.b1", "time", "0000"),
					resource.TestCheckResourceAttr("upcloud_storage_backup.b1", "interval", "daily"),
					resource.TestCheckResourceAttr("upcloud_storage_backup.b1", "retention", "4"),
					testAccCheckStorageBackupRule("s1", &upcloud.BackupRule{
						Time:      "0000",
						Interval:  "daily",
						Retention: 4,
					}),
				),
			},
			{
				Config: `
					resource "upcloud_storage" "s1" {
						title = "addon"
						size = 10
						zone = "pl-waw1"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStorageBackupRule("s1", &upcloud.BackupRule{}),
				),
			},
		},
	})
}

func TestAccUpCloudStorageBackup_withStorageSwitching(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "upcloud_storage" "s1" {
						title = "addon"
						size = 10
						zone = "pl-waw1"
					}

					resource "upcloud_storage" "s2" {
						title = "addon"
						size = 10
						zone = "pl-waw1"
					}

					resource "upcloud_storage_backup" "b1" {
						storage = upcloud_storage.s1.id
						time = "2200"
						interval = "mon"
						retention = 2
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_storage_backup.b1", "time", "2200"),
					resource.TestCheckResourceAttr("upcloud_storage_backup.b1", "interval", "mon"),
					resource.TestCheckResourceAttr("upcloud_storage_backup.b1", "retention", "2"),
					testAccCheckStorageBackupRule("s1", &upcloud.BackupRule{
						Time:      "2200",
						Interval:  "mon",
						Retention: 2,
					}),
				),
			},
			{
				Config: `
					resource "upcloud_storage" "s1" {
						title = "addon"
						size = 10
						zone = "pl-waw1"
					}

					resource "upcloud_storage" "s2" {
						title = "addon"
						size = 10
						zone = "pl-waw1"
					}

					resource "upcloud_storage_backup" "b1" {
						storage = upcloud_storage.s2.id
						time = "0000"
						interval = "daily"
						retention = 4
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_storage_backup.b1", "time", "0000"),
					resource.TestCheckResourceAttr("upcloud_storage_backup.b1", "interval", "daily"),
					resource.TestCheckResourceAttr("upcloud_storage_backup.b1", "retention", "4"),
					testAccCheckStorageBackupRule("s1", &upcloud.BackupRule{}),
					testAccCheckStorageBackupRule("s2", &upcloud.BackupRule{
						Time:      "0000",
						Interval:  "daily",
						Retention: 4,
					}),
				),
			},
		},
	})
}
