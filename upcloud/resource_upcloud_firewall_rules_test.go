package upcloud

import (
	"fmt"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestUpcloudFirewallRules_basic(t *testing.T) {
	var providers []*schema.Provider

	var firewallRules upcloud.FirewallRules
	resourceName := "upcloud_firewall_rules.my_server"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckFirewallRulesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFirewallRulesInstanceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "firewall_rule.#", "1"),
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
				),
			},
		},
	})
}

func TestUpcloudFirewallRules_update(t *testing.T) {
	var providers []*schema.Provider

	var firewallRules upcloud.FirewallRules
	resourceName := "upcloud_firewall_rules.my_server"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckFirewallRulesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFirewallRulesInstanceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "firewall_rule.#", "1"),
					testAccCheckFirewallRulesExists(resourceName, &firewallRules),
				),
			},
			{
				Config: testUpcloudFirewallRulesInstanceConfig_update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "firewall_rule.#", "2"),
					testAccCheckFirewallRulesExists(resourceName, &firewallRules),
				),
			},
		},
	})
}

func TestUpcloudFirewallRules_import(t *testing.T) {
	var providers []*schema.Provider

	var firewallRules upcloud.FirewallRules
	resourceName := "upcloud_firewall_rules.my_server"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckFirewallRulesDestroy,
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
		firewallRules, err := client.GetFirewallRules(&request.GetFirewallRulesRequest{
			ServerUUID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("[WARN] Error listing firewall rules when deleting upcloud firewall rules (%s): %s", rs.Primary.ID, err)
		}

		if len(firewallRules.FirewallRules) != 0 {
			return fmt.Errorf("[WARN] Error  %d firewall rules found against server (%s)", len(firewallRules.FirewallRules), rs.Primary.ID)
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
		latest, err := client.GetFirewallRules(&request.GetFirewallRulesRequest{
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

func testAccCheckUpCloudFirewallRuleAttributes(firewallRules *upcloud.FirewallRules, index int, action, comment, family, icmpType, protocol, direction, destination_address_start, destination_address_end, destination_port_start, destination_port_end, source_address_start, source_address_end, source_port_start, source_port_end string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

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

		if firewallRule.DestinationAddressStart != destination_address_start {
			return fmt.Errorf("Bad destination_address_start, expected (%s), got (%s)", destination_address_start, firewallRule.DestinationAddressStart)
		}

		if firewallRule.DestinationAddressEnd != destination_address_end {
			return fmt.Errorf("Bad destination_address_end, expected (%s), got (%s)", destination_address_end, firewallRule.DestinationAddressEnd)
		}

		if firewallRule.DestinationPortStart != destination_port_start {
			return fmt.Errorf("Bad destination_port_start, expected (%s), got (%s)", destination_port_start, firewallRule.DestinationPortStart)
		}

		if firewallRule.DestinationPortEnd != destination_port_end {
			return fmt.Errorf("Bad destination_port_end, expected (%s), got (%s)", destination_port_end, firewallRule.DestinationPortEnd)
		}

		if firewallRule.SourceAddressStart != source_address_start {
			return fmt.Errorf("Bad source_address_start, expected (%s), got (%s)", source_address_start, firewallRule.SourceAddressStart)
		}

		if firewallRule.SourceAddressEnd != source_address_end {
			return fmt.Errorf("Bad source_address_end, expected (%s), got (%s)", source_address_end, firewallRule.SourceAddressEnd)
		}

		if firewallRule.SourcePortStart != source_port_start {
			return fmt.Errorf("Bad source_port_start, expected (%s), got (%s)", source_port_start, firewallRule.SourcePortStart)
		}

		if firewallRule.SourcePortEnd != source_port_end {
			return fmt.Errorf("Bad source_port_end, expected (%s), got (%s)", source_port_end, firewallRule.SourcePortEnd)
		}

		return nil
	}
}

func testUpcloudFirewallRulesInstanceConfig() string {
	return fmt.Sprintf(`
		resource "upcloud_server" "my_server" {
		  zone     = "fi-hel1"
		  hostname = "debian.example.com"
		  plan     = "1xCPU-2GB"

		  storage_devices {
			action = "create"
			size   = 10
			tier   = "maxiops"
		  }

		  network_interface {
			type = "utility"
		  }

		}

		resource "upcloud_firewall_rules" "my_server" {
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

		}`)
}

func testUpcloudFirewallRulesInstanceConfig_update() string {
	return fmt.Sprintf(`
		resource "upcloud_server" "my_server" {
		  zone     = "fi-hel1"
		  hostname = "debian.example.com"
		  plan     = "1xCPU-2GB"

		  storage_devices {
			action = "create"
			size   = 10
			tier   = "maxiops"
		  }

		  network_interface {
			type = "utility"
		  }

		}

		resource "upcloud_firewall_rules" "my_server" {
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
			source_address_end = "192.168.3.255"
			source_address_start = "192.168.3.1"
		  }

		}`)
}
