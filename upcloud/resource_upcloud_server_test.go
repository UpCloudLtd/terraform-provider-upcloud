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
					size   = 70
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
					size   = 70
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
					size   = 70
					tier   = "maxiops"
				},
			]
		}
`

func Test_serverRestartIsRequired(t *testing.T) {
	type args struct {
		storageDevices []interface{}
	}
	cases := []struct {
		name string
		args args
		want bool
	}{
		{"Server reboot is not required if there's not any valid backup rules", args{
			storageDevices: []interface{}{
				map[string]interface{}{
					"id":          "1",
					"action":      "clone",
					"backup_rule": map[string]interface{}{},
				},
			},
		}, false},
		{"Server reboot is required if there's at least one valid backup rule", args{
			storageDevices: []interface{}{
				map[string]interface{}{
					"id":     "1",
					"action": "clone",
					"backup_rule": map[string]interface{}{
						"interval":  "test-interval",
						"time":      "test-time",
						"retention": "test-retention",
					},
				},
				map[string]interface{}{
					"id":          "2",
					"action":      "clone",
					"backup_rule": map[string]interface{}{},
				},
				map[string]interface{}{
					"id":     "3",
					"action": "clone",
					"backup_rule": map[string]interface{}{
						"interval":  "test-interval",
						"time":      "test-time",
						"retention": "test-retention",
					},
				},
			},
		}, true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := serverRestartIsRequired(tt.args.storageDevices); got != tt.want {
				t.Errorf("serverRestartIsRequired() = %v, want %v", got, tt.want)
			}
		})
	}
}
