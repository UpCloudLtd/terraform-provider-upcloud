package upcloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strconv"
	"testing"
)

func TestAccDataSourceUpCloudZones_basic(t *testing.T) {
	var providers []*schema.Provider

	resourceName := "data.upcloud_zones.empty"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudZonesConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceUpCloudZonesCheck(resourceName),
				),
			},
		},
	})
}

func TestAccDataSourceUpCloudZones_public(t *testing.T) {
	var providers []*schema.Provider

	filterType := "public"
	resourceName := fmt.Sprintf("data.upcloud_zones.%s", filterType)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudZonesConfig_filter(filterType),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceUpCloudZonesCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filter_type", filterType),
					resource.TestCheckResourceAttr(resourceName, "zone_ids.#", "9"),
				),
			},
		},
	})
}

func TestAccDataSourceUpCloudZones_private(t *testing.T) {
	var providers []*schema.Provider

	filterType := "private"
	resourceName := fmt.Sprintf("data.upcloud_zones.%s", filterType)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudZonesConfig_filter(filterType),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceUpCloudZonesCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filter_type", filterType),
					resource.TestCheckResourceAttr(resourceName, "zone_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceUpCloudZones_all(t *testing.T) {
	var providers []*schema.Provider

	filterType := "all"
	resourceName := fmt.Sprintf("data.upcloud_zones.%s", filterType)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudZonesConfig_filter(filterType),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceUpCloudZonesCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filter_type", filterType),
					resource.TestCheckResourceAttr(resourceName, "zone_ids.#", "10"),
				),
			},
		},
	})
}

func testAccDataSourceUpCloudZonesCheck(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		zoneIds, zoneIdsOk := rs.Primary.Attributes["zone_ids.#"]

		if !zoneIdsOk {
			return fmt.Errorf("zone_ids attribute is missing.")
		}

		zoneIdsQuantity, err := strconv.Atoi(zoneIds)

		if err != nil {
			return fmt.Errorf("error parsing names (%s) into integer: %s", zoneIds, err)
		}

		if zoneIdsQuantity == 0 {
			return fmt.Errorf("No zone ids found, this is probably a bug.")
		}

		return nil
	}
}

func testAccDataSourceUpCloudZonesConfig_empty() string {
	return `
data "upcloud_zones" "empty" {}
`
}

func testAccDataSourceUpCloudZonesConfig_filter(filterType string) string {
	return fmt.Sprintf(`
data "upcloud_zones" "%[1]s" {
	filter_type = "%[1]s"
}`, filterType)
}
