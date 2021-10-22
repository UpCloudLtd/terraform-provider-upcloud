package upcloud

import (
	"fmt"
	"testing"

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

						properties {
							public_access = true
							ip_filter = ["10.0.0.1/32"]
						}
					}`, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceIdentifier, "name", rName),
					resource.TestCheckResourceAttr(resourceIdentifier, "plan", "1x1xCPU-2GB-25GB"),
					resource.TestCheckResourceAttr(resourceIdentifier, "title", "testtitle"),
					resource.TestCheckResourceAttr(resourceIdentifier, "zone", "fi-hel1"),
					resource.TestCheckResourceAttr(resourceIdentifier, "powered", "true"),
					resource.TestCheckResourceAttr(resourceIdentifier, "properties.0.ip_filter.0", "10.0.0.1/32"),
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

						properties {
							ip_filter = []
						}
					}`, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceIdentifier, "title", "testtitle modified"),
					resource.TestCheckResourceAttr(resourceIdentifier, "properties.0.public_access", "false"),
					resource.TestCheckResourceAttr(resourceIdentifier, "properties.0.ip_filter.#", "0"),
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
						powered = "true"
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
