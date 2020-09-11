package upcloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strconv"
	"testing"
)

func TestAccUpCloudHosts_basic(t *testing.T) {
	var providers []*schema.Provider

	resourceName := "data.upcloud_hosts.empty"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudHostsConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceUpCloudHostsCheck(resourceName),
				),
			},
		},
	})
}

func testAccDataSourceUpCloudHostsCheck(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		hosts, hostsOk := rs.Primary.Attributes["hosts.#"]

		if !hostsOk {
			return fmt.Errorf("hosts attribute is missing.")
		}

		hostsQuantity, err := strconv.Atoi(hosts)

		if err != nil {
			return fmt.Errorf("error parsing size of hosts (%s) into integer: %s", hosts, err)
		}

		if hostsQuantity == 0 {
			return fmt.Errorf("No hosts found, this is probably a bug.")
		}

		return nil
	}
}

func testAccDataSourceUpCloudHostsConfig_empty() string {
	return `
data "upcloud_hosts" "empty" {}
`
}
