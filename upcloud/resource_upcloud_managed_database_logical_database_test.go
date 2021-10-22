package upcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudManagedDatabasePostgreSQLLogicalDatabase_CreateUpdate(t *testing.T) {
	var providers []*schema.Provider
	rNameManagedDatabase := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	rNameLdb := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourceIdentifierManagedDatabase := fmt.Sprintf("upcloud_managed_database_postgresql.%s", rNameManagedDatabase)
	resourceIdentifierManagedDatabaseLdb := fmt.Sprintf("upcloud_managed_database_logical_database.%s", rNameLdb)
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
					}
					resource "upcloud_managed_database_logical_database" "%[2]s" {
					  service = upcloud_managed_database_postgresql.%[1]s.id
					  name = "%[2]s"
					}

				`, rNameManagedDatabase, rNameLdb),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceIdentifierManagedDatabase, "name", rNameManagedDatabase),
					resource.TestCheckResourceAttrSet(resourceIdentifierManagedDatabase, "service_uri"),
					resource.TestCheckResourceAttr(resourceIdentifierManagedDatabaseLdb, "name", rNameLdb),
				),
			},
		},
	})
}
