package upcloud

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceUpcloudManagedDatabaseValkeySessions(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/data_source_valkey_sessions_s1.tf")

	name := "data.upcloud_managed_database_valkey_sessions.valkey_sessions"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(name, "service"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "sessions.*", map[string]string{
						"query": "info",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(name, "sessions.*", map[string]string{
						"query": "ping",
					}),
				),
			},
		},
	})
}
