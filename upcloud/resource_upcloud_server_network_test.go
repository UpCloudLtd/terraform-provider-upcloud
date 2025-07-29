package upcloud

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

	this := "upcloud_server.this"
	family := "upcloud_server.family"

	var thisIP1, thisIP4, thisIP5 string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(this, "network_interface.#", "5"),
					resource.TestCheckResourceAttr(this, "network_interface.0.index", "1"),
					checkStringDoesNotChange(this, "network_interface.0.ip_address", &thisIP1),
					resource.TestCheckResourceAttr(this, "network_interface.2.index", "3"),
					// Private IP will be re-assigned because it is not specified in the configuration
					resource.TestCheckResourceAttr(this, "network_interface.3.index", "4"),
					checkStringDoesNotChange(this, "network_interface.3.ip_address", &thisIP4),
					resource.TestCheckResourceAttr(this, "network_interface.3.additional_ip_address.#", "0"),
					resource.TestCheckResourceAttr(this, "network_interface.4.index", "5"),
					checkStringDoesNotChange(this, "network_interface.4.ip_address", &thisIP5),
					// IP family change
					resource.TestCheckResourceAttr(family, "network_interface.0.ip_address_family", "IPv4"),
				),
			},
			{
				Config: testDataS2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(this, "network_interface.#", "4"),
					resource.TestCheckResourceAttr(this, "network_interface.0.index", "10"),
					checkStringDoesNotChange(this, "network_interface.0.ip_address", &thisIP1),
					resource.TestCheckResourceAttr(this, "network_interface.1.index", "3"),
					// Private IP will be re-assigned because it is not specified in the configuration
					resource.TestCheckResourceAttr(this, "network_interface.2.index", "4"),
					checkStringDoesNotChange(this, "network_interface.2.ip_address", &thisIP4),
					resource.TestCheckResourceAttr(this, "network_interface.2.additional_ip_address.#", "1"),
					resource.TestCheckResourceAttr(this, "network_interface.3.index", "5"),
					checkStringDoesNotChange(this, "network_interface.3.ip_address", &thisIP5),
					// IP family change
					resource.TestCheckResourceAttr(family, "network_interface.0.ip_address_family", "IPv6"),
				),
			},
			{
				Config: testDataS3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(this, "network_interface.#", "3"),
					resource.TestCheckResourceAttr(this, "network_interface.0.index", "5"),
					checkStringDoesNotChange(this, "network_interface.0.ip_address", &thisIP5),
					resource.TestCheckResourceAttr(this, "network_interface.1.index", "10"),
					checkStringDoesNotChange(this, "network_interface.1.ip_address", &thisIP1),
					resource.TestCheckResourceAttr(this, "network_interface.2.index", "3"),
					// Private IP will be re-assigned because it is not specified in the configuration
				),
			},
		},
	})
}
