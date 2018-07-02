package upcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestUpcloudServer_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudServerInstanceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_server.my-server", "zone"),
					resource.TestCheckResourceAttrSet("upcloud_server.my-server", "hostname"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "zone", "fi-hel1"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "hostname", "debian.example.com"),
				),
			},
		},
	})
}

func testUpcloudServerInstanceConfig() string {
	return fmt.Sprintf(`
		resource "upcloud_server" "my-server" {
			zone     = "fi-hel1"
			hostname = "debian.example.com"
		
		
			storage_devices = [{
				size    = 50
				action  = "clone"
				storage = "01000000-0000-4000-8000-000020030100"
			},
				{
					action  = "attach"
					storage = "01000000-0000-4000-8000-000020010301"
					type    = "cdrom"
				},
				{
					action = "create"
					size   = 700
					tier   = "maxiops"
				},
			]
		}
`)
}

func TestAccServer_changePlan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfigWithSmallServerPlan,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "plan", "1xCPU-2GB"),
				),
			},
			{
				Config: testAccPlanConfigUpdateServerPlan,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "plan", "2xCPU-4GB"),
				),
			},
		},
	})
}

const testAccServerConfigWithSmallServerPlan = `
resource "upcloud_server" "my-server" {
			zone     = "fi-hel1"
			hostname = "debian.example.com"
			plan     = "1xCPU-2GB"
		
		
			storage_devices = [{
				size    = 50
				action  = "clone"
				storage = "01000000-0000-4000-8000-000020030100"
			},
				{
					action  = "attach"
					storage = "01000000-0000-4000-8000-000020010301"
					type    = "cdrom"
				},
				{
					action = "create"
					size   = 700
					tier   = "maxiops"
				},
			]
		}
`

const testAccPlanConfigUpdateServerPlan = `
resource "upcloud_server" "my-server" {
			zone     = "fi-hel1"
			hostname = "debian.example.com"
			plan     = "2xCPU-4GB"
		
		
			storage_devices = [{
				size    = 50
				action  = "clone"
				storage = "01000000-0000-4000-8000-000020030100"
			},
				{
					action  = "attach"
					storage = "01000000-0000-4000-8000-000020010301"
					type    = "cdrom"
				},
				{
					action = "create"
					size   = 700
					tier   = "maxiops"
				},
			]
		}
`
