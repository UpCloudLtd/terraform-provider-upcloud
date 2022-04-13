package upcloud

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudManagedDatabase(t *testing.T) {
	testDataS1, err := ioutil.ReadFile("testdata/upcloud_managed_database/managed_database_s1.tf")
	if err != nil {
		t.Fatal(err)
	}
	testDataS2, err := ioutil.ReadFile("testdata/upcloud_managed_database/managed_database_s2.tf")
	if err != nil {
		t.Fatal(err)
	}

	var providers []*schema.Provider
	pg1Name := "upcloud_managed_database_postgresql.pg1"
	pg2Name := "upcloud_managed_database_postgresql.pg2"
	msql1Name := "upcloud_managed_database_mysql.msql1"
	lgDBName := "upcloud_managed_database_logical_database.logical_db_1"
	userName := "upcloud_managed_database_user.db_user_1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: string(testDataS1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(pg1Name, "name", "pg1"),
					resource.TestCheckResourceAttr(pg1Name, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(pg1Name, "title", "tf-test-pg-1"),
					resource.TestCheckResourceAttr(pg1Name, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(pg1Name, "powered", "true"),
					resource.TestCheckResourceAttr(pg1Name, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(pg1Name, "maintenance_window_dow", "friday"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.ip_filter.0", "10.0.0.1/32"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.version", "13"),
					resource.TestCheckResourceAttr(pg1Name, "type", string(upcloud.ManagedDatabaseServiceTypePostgreSQL)),
					resource.TestCheckResourceAttrSet(pg1Name, "service_uri"),

					resource.TestCheckResourceAttr(pg2Name, "name", "pg2"),
					resource.TestCheckResourceAttr(pg2Name, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(pg2Name, "title", "tf-test-pg-2"),
					resource.TestCheckResourceAttr(pg2Name, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(pg2Name, "powered", "false"),
					resource.TestCheckResourceAttr(pg2Name, "properties.0.version", "13"),

					resource.TestCheckResourceAttr(msql1Name, "name", "msql1"),
					resource.TestCheckResourceAttr(msql1Name, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(msql1Name, "title", "tf-test-msql-1"),
					resource.TestCheckResourceAttr(msql1Name, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(msql1Name, "powered", "true"),

					resource.TestCheckResourceAttr(lgDBName, "name", "tf-test-logical-db-1"),
					resource.TestCheckResourceAttrSet(lgDBName, "service"),

					resource.TestCheckResourceAttr(userName, "username", "somename"),
					resource.TestCheckResourceAttr(userName, "password", "Superpass123"),
					resource.TestCheckResourceAttrSet(userName, "service"),
				),
			},
			{
				Config: string(testDataS2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(pg1Name, "title", "tf-test-updated-pg-1"),
					resource.TestCheckResourceAttr(pg1Name, "maintenance_window_time", "11:00:00"),
					resource.TestCheckResourceAttr(pg1Name, "maintenance_window_dow", "thursday"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.public_access", "false"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.ip_filter.#", "0"),
					resource.TestCheckResourceAttr(pg1Name, "properties.0.version", "14"),
					resource.TestCheckResourceAttr(pg1Name, "powered", "false"),

					resource.TestCheckResourceAttr(pg2Name, "title", "tf-test-updated-pg-2"),
					resource.TestCheckResourceAttr(pg2Name, "powered", "true"),
					resource.TestCheckResourceAttr(pg2Name, "properties.0.version", "14"),

					resource.TestCheckResourceAttr(msql1Name, "title", "tf-test-updated-msql-1"),

					resource.TestCheckResourceAttr(lgDBName, "name", "tf-test-updated-logical-db-1"),

					resource.TestCheckResourceAttr(userName, "password", "Superpass890"),
				),
			},
		},
	})
}

// TODO move the tests below to db utils test file, once the managed database resources are moved under their own package
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
