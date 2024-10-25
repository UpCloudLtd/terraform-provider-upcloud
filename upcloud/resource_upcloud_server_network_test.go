package upcloud

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
