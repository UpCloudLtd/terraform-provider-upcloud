package upcloud

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccUpcloudServerNetwork(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_server/server_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_server/server_s2.tf")

	server1Name := "upcloud_server.server1"

	verifyImportStep := func(name string) resource.TestStep {
		return resource.TestStep{
			Config:            testDataS1,
			ResourceName:      name,
			ImportState:       true,
			ImportStateVerify: false, // IP order is not guaranteed between a creation request and response, so import verification is not feasible
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(server1Name, "network_interface.0.ip_address", "172.102.0.2"),
					resource.TestCheckResourceAttr(server1Name, "network_interface.0.additional_ip_address.#", "1"),
				),
			},
			verifyImportStep(server1Name),
			{
				Config: testDataS2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(server1Name, "network_interface.0.ip_address", "172.102.0.2"),
					resource.TestCheckResourceAttr(server1Name, "network_interface.0.additional_ip_address.#", "4"),
				),
			},
		},
	})
}

func checkStringDoesNotChange(name, key string, expected *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		actual := rs.Primary.Attributes[key]
		if *expected == "" {
			*expected = actual
		} else if actual != *expected {
			return fmt.Errorf(`expected %s to match previous value "%s", got "%s"`, key, *expected, actual)
		}
		return nil
	}
}

func TestAccUpcloudServerInterfaceMatching(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_server/server_ifaces_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_server/server_ifaces_s2.tf")
	testDataS3 := utils.ReadTestDataFile(t, "testdata/upcloud_server/server_ifaces_s3.tf")

	serverName := "upcloud_server.this"

	var ip1, ip3, ip4 string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(serverName, "network_interface.#", "4"),
					resource.TestCheckResourceAttr(serverName, "network_interface.0.index", "1"),
					checkStringDoesNotChange(serverName, "network_interface.0.ip_address", &ip1),
					resource.TestCheckResourceAttr(serverName, "network_interface.2.index", "3"),
					checkStringDoesNotChange(serverName, "network_interface.2.ip_address", &ip3),
					resource.TestCheckResourceAttr(serverName, "network_interface.3.index", "4"),
					checkStringDoesNotChange(serverName, "network_interface.3.ip_address", &ip4),
				),
			},
			{
				Config: testDataS2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(serverName, "network_interface.#", "3"),
					resource.TestCheckResourceAttr(serverName, "network_interface.0.index", "1"),
					checkStringDoesNotChange(serverName, "network_interface.0.ip_address", &ip1),
					resource.TestCheckResourceAttr(serverName, "network_interface.1.index", "3"),
					checkStringDoesNotChange(serverName, "network_interface.1.ip_address", &ip3),
					resource.TestCheckResourceAttr(serverName, "network_interface.2.index", "4"),
					checkStringDoesNotChange(serverName, "network_interface.2.ip_address", &ip4),
				),
			},
			{
				Config: testDataS3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(serverName, "network_interface.#", "3"),
					resource.TestCheckResourceAttr(serverName, "network_interface.0.index", "4"),
					checkStringDoesNotChange(serverName, "network_interface.0.ip_address", &ip4),
					resource.TestCheckResourceAttr(serverName, "network_interface.1.index", "1"),
					checkStringDoesNotChange(serverName, "network_interface.1.ip_address", &ip1),
					resource.TestCheckResourceAttr(serverName, "network_interface.2.index", "3"),
					checkStringDoesNotChange(serverName, "network_interface.2.ip_address", &ip3),
				),
			},
		},
	})
}
