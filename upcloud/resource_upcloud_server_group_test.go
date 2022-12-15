package upcloud

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpCloudServerGroup(t *testing.T) {
	testDataStep1, err := os.ReadFile("testdata/upcloud_server_group/step1.tf")
	if err != nil {
		t.Fatal(err)
	}

	testDataStep2, err := os.ReadFile("testdata/upcloud_server_group/step2.tf")
	if err != nil {
		t.Fatal(err)
	}

	var providers []*schema.Provider

	group1 := "upcloud_server_group.tf_test_1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: string(testDataStep1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(group1, "title", "tf_test_1"),
					resource.TestCheckResourceAttr(group1, "members.#", "1"),
					resource.TestCheckResourceAttr(group1, "labels.%", "3"),
					resource.TestCheckResourceAttr(group1, "anti_affinity", "false"),
				),
			},
			{
				Config: string(testDataStep2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(group1, "title", "tf_test_1_updated"),
					resource.TestCheckResourceAttr(group1, "anti_affinity", "true"),
					resource.TestCheckResourceAttr(group1, "members.#", "0"),
					resource.TestCheckResourceAttr(group1, "labels.%", "2"),
					resource.TestCheckResourceAttr(group1, "labels.key1", "val1"),
				),
			},
		},
	})
}
