package upcloud

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudManagedDatabasePostgreSQL_CreateUpdate(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourceIdentifier := fmt.Sprintf("upcloud_managed_database_postgresql.%s", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "upcloud_managed_database_postgresql" "%[1]s" {
						name = "%[1]s"
						plan = "1x1xCPU-2GB-25GB"
						title = "testtitle"
						zone = "fi-hel1"
						maintenance_window_time = "10:00:00"
  						maintenance_window_dow = "friday"
						properties {
							public_access = true
							ip_filter = ["10.0.0.1/32"]
							version = 13
						}
					}`, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceIdentifier, "name", rName),
					resource.TestCheckResourceAttr(resourceIdentifier, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(resourceIdentifier, "title", "testtitle"),
					resource.TestCheckResourceAttr(resourceIdentifier, "zone", "fi-hel1"),
					resource.TestCheckResourceAttr(resourceIdentifier, "powered", "true"),
					resource.TestCheckResourceAttr(resourceIdentifier, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceIdentifier, "maintenance_window_dow", "friday"),
					resource.TestCheckResourceAttr(resourceIdentifier, "properties.0.ip_filter.0", "10.0.0.1/32"),
					resource.TestCheckResourceAttr(resourceIdentifier, "properties.0.version", "13"),
					resource.TestCheckResourceAttr(resourceIdentifier, "type", string(upcloud.ManagedDatabaseServiceTypePostgreSQL)),
					resource.TestCheckResourceAttrSet(resourceIdentifier, "service_uri"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "upcloud_managed_database_postgresql" "%[1]s" {
						name = "%[1]s"
						plan = "1x1xCPU-2GB-25GB"
						title = "testtitle modified"
						zone = "fi-hel1"
						maintenance_window_time = "11:00:00"
						maintenance_window_dow = "friday"
						properties {
							ip_filter = []
							version = 14
						}
					}`, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceIdentifier, "title", "testtitle modified"),
					resource.TestCheckResourceAttr(resourceIdentifier, "maintenance_window_time", "11:00:00"),
					resource.TestCheckResourceAttr(resourceIdentifier, "maintenance_window_dow", "friday"),
					resource.TestCheckResourceAttr(resourceIdentifier, "properties.0.public_access", "false"),
					resource.TestCheckResourceAttr(resourceIdentifier, "properties.0.ip_filter.#", "0"),
					resource.TestCheckResourceAttr(resourceIdentifier, "properties.0.version", "14"),
				),
			},
		},
	})
}

func TestAccUpcloudManagedDatabasePostgreSQL_CreateAsPoweredOff(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourceIdentifier := fmt.Sprintf("upcloud_managed_database_postgresql.%s", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "upcloud_managed_database_postgresql" "%[1]s" {
						name = "%[1]s"
						plan = "1x1xCPU-2GB-25GB"
						title = "testtitle"
						zone = "fi-hel1"
						powered = false
					}`, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceIdentifier, "name", rName),
					resource.TestCheckResourceAttr(resourceIdentifier, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(resourceIdentifier, "title", "testtitle"),
					resource.TestCheckResourceAttr(resourceIdentifier, "zone", "fi-hel1"),
					resource.TestCheckResourceAttr(resourceIdentifier, "powered", "false"),
					resource.TestCheckResourceAttr(resourceIdentifier, "type", string(upcloud.ManagedDatabaseServiceTypePostgreSQL)),
					resource.TestCheckResourceAttrSet(resourceIdentifier, "service_uri"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "upcloud_managed_database_postgresql" "%[1]s" {
						name = "%[1]s"
						plan = "1x1xCPU-2GB-25GB"
						title = "testtitle"
						zone = "fi-hel1"
						powered = true
					}`, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceIdentifier, "powered", "true"),
				),
			},
		},
	})
}

func TestAccUpcloudManagedDatabaseMySQL_Create(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourceIdentifier := fmt.Sprintf("upcloud_managed_database_mysql.%s", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "upcloud_managed_database_mysql" "%[1]s" {
						name = "%[1]s"
						plan = "1x1xCPU-2GB-25GB"
						title = "testtitle"
						zone = "fi-hel1"
					}`, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceIdentifier, "name", rName),
					resource.TestCheckResourceAttr(resourceIdentifier, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(resourceIdentifier, "title", "testtitle"),
					resource.TestCheckResourceAttr(resourceIdentifier, "zone", "fi-hel1"),
					resource.TestCheckResourceAttr(resourceIdentifier, "powered", "true"),
					resource.TestCheckResourceAttr(resourceIdentifier, "type", string(upcloud.ManagedDatabaseServiceTypeMySQL)),
					resource.TestCheckResourceAttrSet(resourceIdentifier, "service_uri"),
				),
			},
		},
	})
}

// Disable this test for now, as it started causing issues in CI. Need to investigate why and fix it
// func TestAccUpcloudManagedDatabasePostgreSQL_VersionUpgrade(t *testing.T) {
// 	var providers []*schema.Provider

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:          func() { testAccPreCheck(t) },
// 		ProviderFactories: testAccProviderFactories(&providers),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: `
// 					resource "upcloud_managed_database_postgresql" "test_pg" {
// 						name = "testpg"
// 						plan = "1x1xCPU-2GB-25GB"
// 						title = "testversion"
// 						zone = "pl-waw1"
// 						powered = false
// 						properties {
// 							version = 12
// 						}
// 					}`,
// 				Check: resource.TestCheckResourceAttr("upcloud_managed_database_postgresql.test_pg", "properties.0.version", "12"),
// 			},
// 			{
// 				// Check if turning db on and upgrading works
// 				Config: `
// 					resource "upcloud_managed_database_postgresql" "test_pg" {
// 						name = "testpg"
// 						plan = "1x1xCPU-2GB-25GB"
// 						title = "testversion"
// 						zone = "pl-waw1"
// 						powered = true
// 						properties {
// 							version = 13
// 						}
// 					}`,
// 				Check: resource.TestCheckResourceAttr("upcloud_managed_database_postgresql.test_pg", "properties.0.version", "13"),
// 			},
// 			{
// 				// Check if turning db off and upgrading works
// 				Config: `
// 					resource "upcloud_managed_database_postgresql" "test_pg" {
// 						name = "testpg"
// 						plan = "1x1xCPU-2GB-25GB"
// 						title = "testversion"
// 						zone = "pl-waw1"
// 						powered = false
// 						properties {
// 							version = 14
// 						}
// 					}`,
// 				Check: resource.TestCheckResourceAttr("upcloud_managed_database_postgresql.test_pg", "properties.0.version", "14"),
// 			},
// 		},
// 	})
// }

func TestIsManagedDatabaseFullyCreated(t *testing.T) {
	db := &upcloud.ManagedDatabase{
		Backups: make([]upcloud.ManagedDatabaseBackup, 0),
		State:   upcloud.ManagedDatabaseStatePoweroff,
		Users:   make([]upcloud.ManagedDatabaseUser, 0),
	}
	if isManagedDatabaseFullyCreated(db) {
		t.Errorf("isManagedDatabaseFullyCreated failed want false got true %+v", db)
	}

	db.State = upcloud.ManagedDatabaseStateRunning
	db.Backups = append(db.Backups, upcloud.ManagedDatabaseBackup{})
	if isManagedDatabaseFullyCreated(db) {
		t.Errorf("isManagedDatabaseFullyCreated failed want false got true %+v", db)
	}

	db.Users = append(db.Users, upcloud.ManagedDatabaseUser{})
	db.Backups = make([]upcloud.ManagedDatabaseBackup, 0)
	if isManagedDatabaseFullyCreated(db) {
		t.Errorf("isManagedDatabaseFullyCreated failed want false got true %+v", db)
	}

	db.Backups = append(db.Backups, upcloud.ManagedDatabaseBackup{})
	if !isManagedDatabaseFullyCreated(db) {
		t.Errorf("isManagedDatabaseFullyCreated failed want true got false %+v", db)
	}
}

func TestWaitServiceNameToPropagate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	name := "upcloud.com"
	if err := waitServiceNameToPropagate(ctx, name); err != nil {
		t.Errorf("waitServiceNameToPropagate failed %+v", err)
	}
}

func TestWaitServiceNameToPropagateContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()
	name := "upcloud.com"
	if err := waitServiceNameToPropagate(ctx, name); err == nil {
		d, _ := ctx.Deadline()
		t.Errorf("waitServiceNameToPropagate failed didn't timeout before deadline %s", d.Format(time.RFC3339))
	}
}
