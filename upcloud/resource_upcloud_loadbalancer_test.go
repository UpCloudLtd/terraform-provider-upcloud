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
	lbName := "upcloud_loadbalancer.lb"
	dnsName := "upcloud_loadbalancer_resolver.lb_dns_1"
	be1Name := "upcloud_loadbalancer_backend.lb_be_1"
	be2Name := "upcloud_loadbalancer_backend.lb_be_2"
	be1sm1Name := "upcloud_loadbalancer_static_backend_member.lb_be_1_sm_1"
	be1dm1Name := "upcloud_loadbalancer_dynamic_backend_member.lb_be_1_dm_1"
	be2sm1Name := "upcloud_loadbalancer_static_backend_member.lb_be_2_sm_1"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: string(testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(lbName, "plan", "development"),
					resource.TestCheckResourceAttr(lbName, "zone", "fi-hel2"),
					resource.TestCheckResourceAttr(dnsName, "name", "lb-resolver-1-test"),
					resource.TestCheckResourceAttr(dnsName, "nameservers.#", "2"),
					resource.TestCheckResourceAttr(be1Name, "name", "lb-be-1-test"),
					resource.TestCheckResourceAttr(be1Name, "resolver_name", "lb-resolver-1-test"),
					resource.TestCheckResourceAttr(be2Name, "name", "lb-be-2-test"),
					resource.TestCheckResourceAttr(be2Name, "resolver_name", ""),
					resource.TestCheckResourceAttr(be1sm1Name, "backend_name", "lb-be-1-test"),
					resource.TestCheckResourceAttr(be1sm1Name, "name", "lb-be-1-sm-1-test"),
					resource.TestCheckResourceAttr(be1sm1Name, "ip", "10.0.0.10"),
					resource.TestCheckResourceAttr(be1sm1Name, "port", "8000"),
					resource.TestCheckResourceAttr(be1sm1Name, "weight", "100"),
					resource.TestCheckResourceAttr(be1sm1Name, "max_sessions", "1000"),
					resource.TestCheckResourceAttr(be1sm1Name, "enabled", "true"),
					resource.TestCheckResourceAttr(be1dm1Name, "backend_name", "lb-be-1-test"),
					resource.TestCheckResourceAttr(be1dm1Name, "name", "lb-be-1-dm-1-test"),
					resource.TestCheckResourceAttr(be1dm1Name, "weight", "10"),
					resource.TestCheckResourceAttr(be1dm1Name, "max_sessions", "10"),
					resource.TestCheckResourceAttr(be1dm1Name, "enabled", "false"),
					resource.TestCheckResourceAttr(be2sm1Name, "backend_name", "lb-be-2-test"),
				),
			},
		},
	})
}
