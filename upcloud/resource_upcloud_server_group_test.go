package upcloud

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpCloudServerGroup(t *testing.T) {
	testDataStep1 := utils.ReadTestDataFile(t, "testdata/upcloud_server_group/step1.tf")
	testDataStep2 := utils.ReadTestDataFile(t, "testdata/upcloud_server_group/step2.tf")

	var providers []*schema.Provider

	group1 := "upcloud_server_group.tf_test_1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testDataStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(group1, "title", "tf_test_1"),
					resource.TestCheckResourceAttr(group1, "members.#", "1"),
					resource.TestCheckResourceAttr(group1, "labels.%", "3"),
					resource.TestCheckResourceAttr(group1, "anti_affinity_policy", "no"),
				),
			},
			{
				Config: testDataStep2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(group1, "title", "tf_test_1_updated"),
					resource.TestCheckResourceAttr(group1, "anti_affinity_policy", "strict"),
					resource.TestCheckResourceAttr(group1, "members.#", "0"),
					resource.TestCheckResourceAttr(group1, "labels.%", "2"),
					resource.TestCheckResourceAttr(group1, "labels.key1", "val1"),
				),
			},
		},
	})
}
