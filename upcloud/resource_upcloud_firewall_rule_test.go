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
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "destination_address_end"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "destination_address_start"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "destination_port_end"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "destination_port_start"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "direction"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "family"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "icmp_type"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "position"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "protocol"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "source_address_end"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "source_address_start"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "source_port_end"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_rule.my-firewall-rule", "source_port_start"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "action", "accept"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "comment", "Allow SSH from this network"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "destination_address_end", ""),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "destination_address_start", ""),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "destination_port_end", "80"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "destination_port_start", "80"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "direction", "in"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "family", "IPv4"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "icmp_type", ""),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "position", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "protocol", ""),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "source_address_end", ""),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "source_address_start", ""),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "source_port_end", ""),
					resource.TestCheckResourceAttr(
						"upcloud_firewall_rule.my-firewall-rule", "source_port_start", ""),
				),
			},
		},
	})
}

func testUpcloudFirewallRuleInstanceConfig() string {
	return fmt.Sprintf(`
		resource "upcloud_firewall_rule" "my-firewall-rule" {
			server_id                 = "${upcloud_server.test.id}"
			action                    = "accept"
			comment                   = "Allow SSH from this network"
			destination_address_end   = ""
			destination_address_start = ""
			destination_port_end      = "80"
			destination_port_start    = "80"
			direction                 = "in"
			family                    = "IPv4"
			icmp_type                 = ""
			position                  = "1"
			protocol                  = ""
			source_address_end        = ""
			source_address_start      = ""
			source_port_end           = ""
			source_port_start         = ""
		}
`)
}
