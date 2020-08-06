package upcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestUpcloudFirewallRule_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFirewallRuleInstanceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "action"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "comment"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "destination_port_end"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "destination_port_start"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "direction"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "family"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "position"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "protocol"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "source_address_end"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "source_address_start"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "action", "accept"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "comment", "Allow SSH from this network"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "destination_port_end", "22"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "destination_port_start", "22"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "direction", "in"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "family", "IPv4"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "position", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "source_address_end", "192.168.1.255"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "source_address_start", "192.168.1.1"),
				),
			},
		},
	})
}

func testUpcloudFirewallRuleInstanceConfig() string {
	return fmt.Sprintf(`
		resource "upcloud_server" "my-server" {
			zone     = "fi-hel1"
			hostname = "debian.example.com"
			plan     = "1xCPU-2GB"

			storage_devices {
					action = "create"
					size   = 10
					tier   = "maxiops"
			}
		}

		resource "upcloud_firewall_rule" "my-firewall-rule" {
			server_id                 = "${upcloud_server.my-server.id}"
			action                    = "accept"
			comment                   = "Allow SSH from this network"
			destination_address_end   = ""
			destination_address_start = ""
			destination_port_end      = "22"
			destination_port_start    = "22"
			direction                 = "in"
			family                    = "IPv4"
			icmp_type                 = ""
			position                  = "1"
			protocol                  = "tcp"
			source_address_end        = "192.168.1.255"
			source_address_start      = "192.168.1.1"
			source_port_end           = ""
			source_port_start         = ""
		}
`)
}
