package upcloud

import (
	"io/ioutil"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudLoadBalancer(t *testing.T) {
	testData, err := ioutil.ReadFile("testdata/upcloud_loadbalancer/loadbalancer.tf")
	if err != nil {
		t.Fatal(err)
	}
	var providers []*schema.Provider
	name := "upcloud_loadbalancer.lb"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: string(testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "plan", "development"),
					resource.TestCheckResourceAttr(name, "zone", "fi-hel2"),
				),
			},
		},
	})
}
