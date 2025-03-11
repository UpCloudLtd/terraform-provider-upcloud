package upcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	configWithID = `
data "upcloud_zone" "my_zone" {
  id = "uk-lon1"
}`
)

func TestAccDataSourceUpCloudZone_basic(t *testing.T) {
	resourceName := "data.upcloud_zone.my_zone"

	expectedZoneName := "uk-lon1"
	expectedDescription := "London #1"
	expectedPublic := "true"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configWithID,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", expectedZoneName),
					resource.TestCheckResourceAttr(resourceName, "description", expectedDescription),
					resource.TestCheckResourceAttr(resourceName, "public", expectedPublic),
				),
			},
		},
	})
}
