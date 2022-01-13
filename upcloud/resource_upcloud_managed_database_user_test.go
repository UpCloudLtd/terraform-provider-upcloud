package upcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudManagedDatabasePostgreSQLUser_CreateUpdate(t *testing.T) {
	var providers []*schema.Provider
	rNameManagedDatabase := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	rNameUser := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	testPassword := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	testPassword2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourceIdentifierManagedDatabase := fmt.Sprintf("upcloud_managed_database_postgresql.%s", rNameManagedDatabase)
	resourceIdentifierManagedDatabaseUser := fmt.Sprintf("upcloud_managed_database_user.%s", rNameUser)
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
					resource "upcloud_managed_database_user" "%[2]s" {
						service = upcloud_managed_database_postgresql.%[1]s.id
						username = "%[2]s"
						password = "%[3]s"
					}
				`, rNameManagedDatabase, rNameUser, testPassword),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceIdentifierManagedDatabase, "name", rNameManagedDatabase),
					resource.TestCheckResourceAttrSet(resourceIdentifierManagedDatabase, "service_uri"),
					resource.TestCheckResourceAttr(resourceIdentifierManagedDatabaseUser, "username", rNameUser),
					resource.TestCheckResourceAttr(resourceIdentifierManagedDatabaseUser, "password", testPassword),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "upcloud_managed_database_postgresql" "%[1]s" {
						name = "%[1]s"
						plan = "1x1xCPU-2GB-25GB"
						title = "testtitle"
						zone = "fi-hel1"
					}
					resource "upcloud_managed_database_user" "%[2]s" {
						service = upcloud_managed_database_postgresql.%[1]s.id
						username = "%[2]s"
						password = "%[3]s"
					}
				`, rNameManagedDatabase, rNameUser, testPassword2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceIdentifierManagedDatabaseUser, "password", testPassword2),
				),
			},
		},
	})
}
