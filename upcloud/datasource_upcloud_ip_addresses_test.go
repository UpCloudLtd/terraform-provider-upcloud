package upcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceUpCloudIPAddresses_basic(t *testing.T) {
	resourceName := "data.upcloud_ip_addresses.empty"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudIPAddressesConfigEmpty(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceUpCloudIPAddressesCheck(resourceName),
				),
			},
		},
	})
}

func testAccDataSourceUpCloudIPAddressesCheck(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		_, ipAddressesOk := rs.Primary.Attributes["addresses.#"]

		if !ipAddressesOk {
			return fmt.Errorf("addresses attribute is missing")
		}

		return nil
	}
}

func testAccDataSourceUpCloudIPAddressesConfigEmpty() string {
	return `
data "upcloud_ip_addresses" "empty" {}
`
}
