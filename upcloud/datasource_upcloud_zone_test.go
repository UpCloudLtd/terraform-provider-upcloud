package upcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUpCloudZone_basic(t *testing.T) {
	resourceName := "data.upcloud_zone.my_zone"

	expectedZoneName := "uk-lon1"
	expectedDescription := "London #1"
	expectedPublic := "true"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudZoneConfig(expectedZoneName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "name", expectedZoneName),
					resource.TestCheckResourceAttr(
						resourceName, "description", expectedDescription),
					resource.TestCheckResourceAttr(
						resourceName, "public", expectedPublic),
				),
			},
		},
	})
}

func testAccDataSourceUpCloudZoneConfig(zoneName string) string {
	return fmt.Sprintf(`
data "upcloud_zone" "my_zone" {
  name = "%s"
}`, zoneName)
}
