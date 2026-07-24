package servertests

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceUpCloudServer(t *testing.T) {
	resourceName := "upcloud_server.test"
	dataSourceName := "data.upcloud_server.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceServerConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "hostname", resourceName, "hostname"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone", resourceName, "zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "plan", resourceName, "plan"),
					resource.TestCheckResourceAttr(dataSourceName, "state", "started"),
					resource.TestCheckResourceAttrSet(dataSourceName, "network_interface.0.ip_address"),
					resource.TestCheckResourceAttrSet(dataSourceName, "network_interface.0.mac_address"),
					resource.TestCheckResourceAttr(dataSourceName, "network_interface.0.type", "utility"),
				),
			},
		},
	})
}

func testAccDataSourceServerConfig() string {
	return fmt.Sprintf(`
resource "upcloud_server" "test" {
  hostname = "tf-acc-test-server-datasource"
  zone     = "fi-hel1"
  plan     = "1xCPU-1GB"

  template {
    storage = "%s"
    size    = 10
  }

  network_interface {
    type = "utility"
  }
}

data "upcloud_server" "test" {
  id = upcloud_server.test.id
}
`, upcloud.DebianTemplateUUID)
}
