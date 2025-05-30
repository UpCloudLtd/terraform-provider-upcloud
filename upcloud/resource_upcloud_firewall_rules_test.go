package upcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const firewallRulesResourceName = "upcloud_firewall_rules.my_rule"

func TestUpcloudFirewallRules_basic(t *testing.T) {
	var firewallRules upcloud.FirewallRules
	resourceName := firewallRulesResourceName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckFirewallRulesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFirewallRulesInstanceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "firewall_rule.#", "2"),
					testAccCheckFirewallRulesExists(resourceName, &firewallRules),
					testAccCheckUpCloudFirewallRuleAttributes(&firewallRules, 0, "accept",
						"Allow SSH from this network",
						"IPv4",
						"",
						"tcp",
						"in",
						"",
						"",
						"22",
						"22",
						"192.168.1.1",
						"192.168.1.255",
						"",
						""),
					testAccCheckUpCloudFirewallRuleAttributes(&firewallRules, 1, "drop",
						"Drop all connection from ip range",
						"IPv4",
						"",
						"",
						"in",
						"",
						"",
						"",
						"",
						"192.168.1.1",
						"192.168.1.255",
						"",
						""),
				),
			},
		},
	})
}

func TestUpcloudFirewallRules_update(t *testing.T) {
	var firewallRules upcloud.FirewallRules
	resourceName := "upcloud_firewall_rules.my_rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckFirewallRulesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFirewallRulesInstanceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "firewall_rule.#", "2"),
					testAccCheckFirewallRulesExists(resourceName, &firewallRules),
				),
			},
			{
				Config: testUpcloudFirewallRulesInstanceConfigUpdate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "firewall_rule.#", "2"),
					testAccCheckFirewallRulesExists(resourceName, &firewallRules),
				),
			},
		},
	})
}

func TestUpcloudFirewallRules_import(t *testing.T) {
	var firewallRules upcloud.FirewallRules
	resourceName := firewallRulesResourceName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckFirewallRulesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFirewallRulesInstanceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRulesExists(resourceName, &firewallRules),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckFirewallRulesDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "upcloud_firewall_rules" {
			continue
		}

		client := testAccProvider.Meta().(*service.Service)

		_, err := client.GetFirewallRules(context.Background(), &request.GetFirewallRulesRequest{
			ServerUUID: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf(
				"Error firewall rules still exists after deletion for server (%s)",
				rs.Primary.ID,
			)
		}
	}

	return nil
}

func testAccCheckFirewallRulesExists(resourceName string, firewallRules *upcloud.FirewallRules) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Look for the full resource name and error if not found
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// The provider has not set the ID for the resource
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Firewall ID is set")
		}

		client := testAccProvider.Meta().(*service.Service)
		latest, err := client.GetFirewallRules(context.Background(), &request.GetFirewallRulesRequest{
			ServerUUID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		// Update the reference the remote located storage
		*firewallRules = *latest

		return nil
	}
}

func testAccCheckUpCloudFirewallRuleAttributes(firewallRules *upcloud.FirewallRules, index int, action, comment, family, icmpType, protocol, direction, destinationAddressStart, destinationAddressEnd, destinationPortStart, destinationPortEnd, sourceAddressStart, sourceAddressEnd, sourcePortStart, sourcePortEnd string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		firewallRule := firewallRules.FirewallRules[index]

		if firewallRule.Action != action {
			return fmt.Errorf("Bad action, expected (%s), got (%s)", action, firewallRule.Action)
		}

		if firewallRule.Comment != comment {
			return fmt.Errorf("Bad comment, expected (%s), got (%s)", comment, firewallRule.Comment)
		}

		if firewallRule.Family != family {
			return fmt.Errorf("Bad family, expected (%s), got (%s)", family, firewallRule.Family)
		}

		if firewallRule.ICMPType != icmpType {
			return fmt.Errorf("Bad icmpType, expected (%s), got (%s)", icmpType, firewallRule.ICMPType)
		}

		if firewallRule.Protocol != protocol {
			return fmt.Errorf("Bad protocol, expected (%s), got (%s)", protocol, firewallRule.Protocol)
		}

		if firewallRule.Direction != direction {
			return fmt.Errorf("Bad direction, expected (%s), got (%s)", direction, firewallRule.Direction)
		}

		if firewallRule.DestinationAddressStart != destinationAddressStart {
			return fmt.Errorf("Bad destination_address_start, expected (%s), got (%s)", destinationAddressStart, firewallRule.DestinationAddressStart)
		}

		if firewallRule.DestinationAddressEnd != destinationAddressEnd {
			return fmt.Errorf("Bad destination_address_end, expected (%s), got (%s)", destinationAddressEnd, firewallRule.DestinationAddressEnd)
		}

		if firewallRule.DestinationPortStart != destinationPortStart {
			return fmt.Errorf("Bad destination_port_start, expected (%s), got (%s)", destinationPortStart, firewallRule.DestinationPortStart)
		}

		if firewallRule.DestinationPortEnd != destinationPortEnd {
			return fmt.Errorf("Bad destination_port_end, expected (%s), got (%s)", destinationPortEnd, firewallRule.DestinationPortEnd)
		}

		if firewallRule.SourceAddressStart != sourceAddressStart {
			return fmt.Errorf("Bad source_address_start, expected (%s), got (%s)", sourceAddressStart, firewallRule.SourceAddressStart)
		}

		if firewallRule.SourceAddressEnd != sourceAddressEnd {
			return fmt.Errorf("Bad source_address_end, expected (%s), got (%s)", sourceAddressEnd, firewallRule.SourceAddressEnd)
		}

		if firewallRule.SourcePortStart != sourcePortStart {
			return fmt.Errorf("Bad source_port_start, expected (%s), got (%s)", sourcePortStart, firewallRule.SourcePortStart)
		}

		if firewallRule.SourcePortEnd != sourcePortEnd {
			return fmt.Errorf("Bad source_port_end, expected (%s), got (%s)", sourcePortEnd, firewallRule.SourcePortEnd)
		}

		return nil
	}
}

func testUpcloudFirewallRulesInstanceConfig() string {
	return fmt.Sprintf(`
		resource "upcloud_server" "my_server" {
			zone     = "fi-hel1"
			hostname = "tf-acc-test-firewall-rules-vm"
			plan     = "1xCPU-2GB"
			metadata = true

			template {
				storage = "%s"
				size = 10
			}

			network_interface {
				type = "utility"
			}

		}

		resource "upcloud_firewall_rules" "my_rule" {
			server_id = upcloud_server.my_server.id

			firewall_rule {
				action = "accept"
				comment = "Allow SSH from this network"
				destination_address_end = ""
				destination_address_start = ""
				destination_port_end = 22
				destination_port_start = 22
				direction = "in"
				family = "IPv4"
				icmp_type = ""
				protocol = "tcp"
				source_address_end = "192.168.1.255"
				source_address_start = "192.168.1.1"
			}

			firewall_rule {
				action = "drop"
				comment = "Drop all connection from ip range"
				direction = "in"
				family = "IPv4"
				source_address_end = "192.168.1.255"
				source_address_start = "192.168.1.1"
			}
		}`, debianTemplateUUID)
}

func testUpcloudFirewallRulesInstanceConfigUpdate() string {
	return fmt.Sprintf(`
		resource "upcloud_server" "my_server" {
			zone     = "fi-hel1"
			hostname = "tf-acc-test-firewall-rules-vm"
			plan     = "1xCPU-2GB"
			metadata = true

			template {
				storage = "%s"
				size = 10
			}

			network_interface {
				type = "utility"
			}

		}

		resource "upcloud_firewall_rules" "my_rule" {
			server_id = upcloud_server.my_server.id

			firewall_rule {
				action = "accept"
				comment = "Allow SSH from this network"
				destination_address_end = ""
				destination_address_start = ""
				destination_port_end = "22"
				destination_port_start = 22
				direction = "in"
				family = "IPv4"
				icmp_type = ""
				protocol = "tcp"
				source_address_end = "192.168.1.255"
				source_address_start = "192.168.1.1"
			}

			firewall_rule {
				action = "accept"
				comment = "Allow SSH from this network"
				destination_address_end = ""
				destination_address_start = ""
				destination_port_end = 22
				destination_port_start = "22"
				direction = "in"
				family = "IPv4"
				icmp_type = ""
				protocol = "tcp"
				source_address_end = "192.168.3.255"
				source_address_start = "192.168.3.1"
			}

		}`, debianTemplateUUID)
}
