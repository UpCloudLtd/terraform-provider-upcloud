package cloudtests

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUpCloudHosts_basic(t *testing.T) {
	resourceName := "data.upcloud_hosts.empty"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudHostsConfigEmpty(),
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
			return fmt.Errorf("hosts attribute is missing")
		}

		hostsQuantity, err := strconv.Atoi(hosts)
		if err != nil {
			return fmt.Errorf("error parsing size of hosts (%s) into integer: %s", hosts, err)
		}

		if hostsQuantity != 0 {
			return fmt.Errorf("some hosts have been found, expecting no hosts")
		}

		return nil
	}
}

func testAccDataSourceUpCloudHostsConfigEmpty() string {
	return `
data "upcloud_hosts" "empty" {}
`
}
