package upcloud

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func withPrefixDB(text string) string {
	return fmt.Sprintf("tf-acc-test-db-%s", text)
}

func TestAccUpcloudManagedDatabase(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/managed_database_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/managed_database_s2.tf")

	pg1Name := "upcloud_managed_database_postgresql.pg1"
	pg2Name := "upcloud_managed_database_postgresql.pg2"
	msql1Name := "upcloud_managed_database_mysql.msql1"
	lgDBName := "upcloud_managed_database_logical_database.logical_db_1"
	userName1 := "upcloud_managed_database_user.db_user_1"
	userName2 := "upcloud_managed_database_user.db_user_2"
	userName4 := "upcloud_managed_database_user.db_user_4"
	userName5 := "upcloud_managed_database_user.db_user_5"
	valkeyName := "upcloud_managed_database_valkey.v1"

	verifyImportStep := func(name string) resource.TestStep {
		return resource.TestStep{
			Config:                  testDataS1,
			ResourceName:            name,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"properties.0.admin_password", "properties.0.admin_username", "state"}, // credentials only provided on creation, not available on subsequent requests like import
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(pg1Name, "additional_disk_space_gib", "20"),
					resource.TestCheckResourceAttr(pg1Name, "name", withPrefixDB("pg-1")),
					resource.TestCheckResourceAttr(pg1Name, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(pg1Name, "title", withPrefixDB("pg-1")),
					resource.TestCheckResourceAttr(pg1Name, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(pg1Name, "powered", "true"),
					// Uncomment these here and in testdata, when the fields can be modified together with powered field.
					// resource.TestCheckResourceAttr(pg1Name, "maintenance_window_time", "10:00:00"),
					// resource.TestCheckResourceAttr(pg1Name, "maintenance_window_dow", "friday"),
					// resource.TestCheckResourceAttr(pg1Name, "properties.0.ip_filter.0", "10.0.0.1/32"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.version", "16"),
					resource.TestCheckResourceAttr(pg1Name, "type", string(upcloud.ManagedDatabaseServiceTypePostgreSQL)),
					resource.TestCheckResourceAttrSet(pg1Name, "service_uri"),
					resource.TestCheckResourceAttr(pg1Name, "network.#", "0"),

					resource.TestCheckResourceAttr(pg2Name, "additional_disk_space_gib", "0"),
					resource.TestCheckResourceAttr(pg2Name, "name", withPrefixDB("pg-2")),
					resource.TestCheckResourceAttr(pg2Name, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(pg2Name, "title", withPrefixDB("pg-2")),
					resource.TestCheckResourceAttr(pg2Name, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(pg2Name, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(pg2Name, "maintenance_window_dow", "friday"),
					resource.TestCheckResourceAttr(pg2Name, "properties.0.ip_filter.0", "10.0.0.1/32"),
					resource.TestCheckResourceAttr(pg2Name, "powered", "true"),
					resource.TestCheckResourceAttr(pg2Name, "properties.0.version", "17"),
					resource.TestCheckResourceAttr(pg2Name, "network.#", "1"),

					resource.TestCheckResourceAttr(msql1Name, "additional_disk_space_gib", "10"),
					resource.TestCheckResourceAttr(msql1Name, "name", withPrefixDB("mysql-1")),
					resource.TestCheckResourceAttr(msql1Name, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(msql1Name, "title", withPrefixDB("mysql-1")),
					resource.TestCheckResourceAttr(msql1Name, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(msql1Name, "powered", "true"),
					resource.TestCheckResourceAttr(msql1Name, "network.#", "0"),

					resource.TestCheckResourceAttr(lgDBName, "name", withPrefixDB("logical-db-1")),
					resource.TestCheckResourceAttrSet(lgDBName, "service"),

					resource.TestCheckResourceAttr(userName1, "username", "somename"),
					resource.TestCheckResourceAttr(userName1, "password", "Superpass123"),
					resource.TestCheckResourceAttr(userName1, "authentication", "mysql_native_password"),
					resource.TestCheckResourceAttrSet(userName1, "service"),

					resource.TestCheckResourceAttr(userName2, "pg_access_control.0.allow_replication", "false"),

					resource.TestCheckResourceAttr(userName5, "valkey_access_control.0.categories.0", "+@all"),
					resource.TestCheckResourceAttr(userName5, "valkey_access_control.0.channels.0", "*"),
					resource.TestCheckResourceAttr(userName5, "valkey_access_control.0.commands.#", "3"),
					resource.TestCheckResourceAttr(userName5, "valkey_access_control.0.keys.0", "key_*"),

					resource.TestCheckResourceAttr(userName4, "opensearch_access_control.0.rules.0.index", ".opensearch-observability"),
					resource.TestCheckResourceAttr(userName4, "opensearch_access_control.0.rules.0.permission", "admin"),

					resource.TestCheckResourceAttr(valkeyName, "name", withPrefixDB("valkey-1")),
					resource.TestCheckResourceAttr(valkeyName, "plan", "1x1xCPU-2GB"),
					resource.TestCheckResourceAttr(valkeyName, "title", withPrefixDB("valkey-1")),
					resource.TestCheckResourceAttr(valkeyName, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(valkeyName, "powered", "true"),
					resource.TestCheckResourceAttr(valkeyName, "network.#", "1"),
				),
			},
			verifyImportStep(pg1Name),
			verifyImportStep(pg2Name),
			verifyImportStep(msql1Name),
			verifyImportStep(valkeyName),
			verifyImportStep(lgDBName),
			verifyImportStep(userName1),
			verifyImportStep(userName2),
			verifyImportStep(userName4),
			verifyImportStep(userName5),
			{
				Config: testDataS2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(pg1Name, "additional_disk_space_gib", "40"),
					resource.TestCheckResourceAttr(pg1Name, "title", withPrefixDB("pg-1-updated")),
					// Uncomment these here and in testdata, when the fields can be modified together with powered field.
					// resource.TestCheckResourceAttr(pg1Name, "maintenance_window_time", "11:00:00"),
					// resource.TestCheckResourceAttr(pg1Name, "maintenance_window_dow", "thursday"),
					// resource.TestCheckResourceAttr(pg1Name, "properties.0.public_access", "false"),
					// resource.TestCheckResourceAttr(pg1Name, "properties.0.ip_filter.#", "0"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.version", "16"),
					resource.TestCheckResourceAttr(pg1Name, "powered", "false"),

					resource.TestCheckResourceAttr(pg2Name, "additional_disk_space_gib", "20"),
					resource.TestCheckResourceAttr(pg2Name, "title", withPrefixDB("pg-2-updated")),
					resource.TestCheckResourceAttr(pg2Name, "maintenance_window_time", "11:00:00"),
					resource.TestCheckResourceAttr(pg2Name, "maintenance_window_dow", "thursday"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.public_access", "false"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.ip_filter.#", "0"),
					resource.TestCheckResourceAttr(pg2Name, "powered", "true"),
					resource.TestCheckResourceAttr(pg2Name, "properties.0.version", "17"),
					resource.TestCheckResourceAttr(pg2Name, "network.#", "1"),

					resource.TestCheckResourceAttr(msql1Name, "additional_disk_space_gib", "0"),
					resource.TestCheckResourceAttr(msql1Name, "title", withPrefixDB("mysql-1-updated")),
					resource.TestCheckResourceAttr(msql1Name, "network.#", "1"),

					resource.TestCheckResourceAttr(lgDBName, "name", withPrefixDB("logical-db-1-updated")),

					resource.TestCheckResourceAttr(userName1, "password", "Superpass890"),
					resource.TestCheckResourceAttr(userName1, "authentication", "caching_sha2_password"),

					resource.TestCheckResourceAttr(userName2, "pg_access_control.0.allow_replication", "true"),

					resource.TestCheckResourceAttr(userName5, "valkey_access_control.0.categories.#", "0"),
					resource.TestCheckResourceAttr(userName5, "valkey_access_control.0.channels.#", "0"),
					resource.TestCheckResourceAttr(userName5, "valkey_access_control.0.commands.#", "0"),
					resource.TestCheckResourceAttr(userName5, "valkey_access_control.0.keys.0", "key*"),

					resource.TestCheckResourceAttr(valkeyName, "name", withPrefixDB("valkey-1")),
					resource.TestCheckResourceAttr(valkeyName, "plan", "1x1xCPU-2GB"),
					resource.TestCheckResourceAttr(valkeyName, "title", withPrefixDB("valkey-1-updated")),
					resource.TestCheckResourceAttr(valkeyName, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(valkeyName, "powered", "true"),
				),
			},
		},
	})
}

func TestAccUpcloudManagedDatabase_terminationProtection(t *testing.T) {
	testdata := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/termination_protection.tf")
	db := "upcloud_managed_database_mysql.this.0"

	variables := func(dbCount int32, powered bool, termination_protection bool) map[string]config.Variable {
		return map[string]config.Variable{
			"db_count":               config.IntegerVariable(dbCount),
			"powered":                config.BoolVariable(powered),
			"termination_protection": config.BoolVariable(termination_protection),
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testdata,
				ConfigVariables: variables(1, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(db, "powered", "true"),
					resource.TestCheckResourceAttr(db, "termination_protection", "true"),
				),
			},
			// Powering off the service should fail as termination protection is enabled
			{
				Config:          testdata,
				ConfigVariables: variables(1, false, true),
				ExpectError:     regexp.MustCompile("Service state cannot be updated, termination protection is enabled."),
			},
			// Deleting the service should fail as termination protection is enabled
			{
				Config:          testdata,
				ConfigVariables: variables(0, true, true),
				ExpectError:     regexp.MustCompile("Service cannot be deleted, termination protection is enabled."),
			},
			// Disable termination protection and power off the service
			{
				Config:          testdata,
				ConfigVariables: variables(1, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(db, "powered", "false"),
					resource.TestCheckResourceAttr(db, "termination_protection", "false"),
				),
			},
			// Deleting the service should succeed as termination protection is disabled
			{
				Config:          testdata,
				ConfigVariables: variables(0, false, false),
			},
		},
	})
}
