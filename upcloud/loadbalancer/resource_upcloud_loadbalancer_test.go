package networkpeeringtests

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func peeringIDFunc(name string) func(state *terraform.State) (string, error) {
	return func(state *terraform.State) (string, error) {
		resourceState := state.RootModule().Resources[name]
		return resourceState.Primary.ID, nil
	}
}

func TestAccUpCloudNetworkPeering(t *testing.T) {
	peering0 := "upcloud_network_peering.this.0"

	testDataStep1 := utils.ReadTestDataFile(t, "../testdata/upcloud_network_peering/network_peering_s1.tf")
	testDataStep2 := utils.ReadTestDataFile(t, "../testdata/upcloud_network_peering/network_peering_s2.tf")
	testDataStep3 := utils.ReadTestDataFile(t, "../testdata/upcloud_network_peering/network_peering_s3.tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(peering0, "name", "tf-acc-test-peering-peering-0-to-1"),
					resource.TestCheckResourceAttr(peering0, "labels.%", "2"),
				),
			},
			{
				Config:            testDataStep1,
				ResourceName:      "upcloud_network_peering.this[0]",
				ImportState:       true,
				ImportStateIdFunc: peeringIDFunc(peering0),
				ImportStateVerify: true,
			},
			{
				Config: testDataStep2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(peering0, "labels.%", "1"),
					resource.TestCheckResourceAttr("upcloud_network_peering.this.1", "labels.%", "1"),
				),
			},
			{
				Config: testDataStep3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(peering0, "name", "tf-acc-test-peering-peering-0-to-1"),
				),
			},
		},
	})
}
