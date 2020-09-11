package upcloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"testing"
)

func TestAccDataSourceUpCloudZone_basic(t *testing.T) {
	var providers []*schema.Provider

	resourceName := "data.upcloud_zone.my_zone"

	expectedZoneName := "uk-lon1"
	expectedDescription := "London #1"
	expectedPublic := "true"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
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
