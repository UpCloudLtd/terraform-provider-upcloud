package database

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceUpcloudManagedDatabasePostgreSQLSessions(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "../testdata/upcloud_managed_database/data_source_postgresql_sessions_s1.tf")

	name := "data.upcloud_managed_database_postgresql_sessions.postgresql_sessions"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(name, "service"),
					resource.TestCheckResourceAttrSet(name, "sessions.#"),
				),
			},
		},
	})
}
