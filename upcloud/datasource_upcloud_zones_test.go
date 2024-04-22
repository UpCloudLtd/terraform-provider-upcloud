package upcloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	availablePublicZones = 13
	allFilter            = "all"
	publicFilter         = "public"
	privateFilter        = "private"
)

func TestAccDataSourceUpCloudZones_default(t *testing.T) {
	resourceName := "data.upcloud_zones.empty"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudZonesConfigEmpty(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceUpCloudZonesCheck(resourceName, availablePublicZones),
				),
			},
		},
	})
}

func TestAccDataSourceUpCloudZones_public(t *testing.T) {
	filterType := publicFilter
	resourceName := fmt.Sprintf("data.upcloud_zones.%s", filterType)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudZonesConfigFilter(filterType),
				Check: resource.ComposeTestCheckFunc(

					testAccDataSourceUpCloudZonesCheck(resourceName, availablePublicZones),
					resource.TestCheckResourceAttr(resourceName, "filter_type", filterType),
				),
			},
		},
	})
}

func TestAccDataSourceUpCloudZones_private(t *testing.T) {
	filterType := privateFilter
	resourceName := fmt.Sprintf("data.upcloud_zones.%s", filterType)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudZonesConfigFilter(filterType),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceUpCloudZonesCheck(resourceName, 0),
					resource.TestCheckResourceAttr(resourceName, "filter_type", filterType),
				),
			},
		},
	})
}

func TestAccDataSourceUpCloudZones_all(t *testing.T) {
	filterType := allFilter
	resourceName := fmt.Sprintf("data.upcloud_zones.%s", filterType)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudZonesConfigFilter(filterType),
				Check: resource.ComposeTestCheckFunc(

					testAccDataSourceUpCloudZonesCheck(resourceName, availablePublicZones),
					resource.TestCheckResourceAttr(resourceName, "filter_type", filterType),
				),
			},
		},
	})
}

func testAccDataSourceUpCloudZonesCheck(resourceName string, expectedResources int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		zoneIDs, zoneIDsOk := rs.Primary.Attributes["zone_ids.#"]

		if !zoneIDsOk {
			return fmt.Errorf("zone_ids attribute is missing")
		}

		zoneIDsQuantity, err := strconv.Atoi(zoneIDs)
		if err != nil {
			return fmt.Errorf("error parsing names (%s) into integer: %s", zoneIDs, err)
		}

		if zoneIDsQuantity != expectedResources {
			return fmt.Errorf("unexpected number of resource (%v), expected %v",
				zoneIDsQuantity, expectedResources)
		}

		return nil
	}
}

func testAccDataSourceUpCloudZonesConfigEmpty() string {
	return `
data "upcloud_zones" "empty" {}
`
}

func testAccDataSourceUpCloudZonesConfigFilter(filterType string) string {
	return fmt.Sprintf(`
data "upcloud_zones" "%[1]s" {
	filter_type = "%[1]s"
}`, filterType)
}
