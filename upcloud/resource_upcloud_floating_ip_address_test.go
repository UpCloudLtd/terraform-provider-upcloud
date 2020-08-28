package upcloud

import (
	"fmt"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudFloatingIPAddress_basic(t *testing.T) {
	var providers []*schema.Provider

	resourceName := "upcloud_floating_ip_address.my_floating_ip"
	expectedMacAddress := ""
	expectedZone := "fi-hel1"
	expectedFamily := "IPv4"
	expectedAccess := "public"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFloatingIPAddressBasicConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mac_address", expectedMacAddress),
					resource.TestCheckResourceAttr(resourceName, "zone", expectedZone),
					resource.TestCheckResourceAttr(resourceName, "family", expectedFamily),
					resource.TestCheckResourceAttr(resourceName, "access", expectedAccess),
					resource.TestCheckResourceAttrSet(resourceName, "ip_address"),
				),
			},
		},
	})
}

func TestAccUpcloudFloatingIPAddress_create_with_server(t *testing.T) {
	var providers []*schema.Provider

	serverResourceName := "upcloud_server.my_server"
	floatingIPResourceName := "upcloud_floating_ip_address.my_floating_ip"
	expectedZone := "fi-hel1"
	expectedFamily := "IPv4"
	expectedAccess := "public"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFloatingIPAddressCreateWithServerConfig([]string{"my_server"}, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(floatingIPResourceName, "zone", expectedZone),
					resource.TestCheckResourceAttr(floatingIPResourceName, "family", expectedFamily),
					resource.TestCheckResourceAttr(floatingIPResourceName, "access", expectedAccess),
					resource.TestCheckResourceAttrSet(floatingIPResourceName, "ip_address"),
					resource.TestCheckResourceAttrSet(floatingIPResourceName, "mac_address"),
					testAccCheckFloatingIP(floatingIPResourceName, serverResourceName),
				),
			},
		},
	})
}

func TestAccUpcloudFloatingIPAddress_switch_between_servers(t *testing.T) {
	var providers []*schema.Provider

	firstServerResourceName := "upcloud_server.my_first_server"
	secondServerResourceName := "upcloud_server.my_second_server"
	floatingIPResourceName := "upcloud_floating_ip_address.my_floating_ip"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFloatingIPAddressCreateWithServerConfig([]string{"my_first_server", "my_second_server"}, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(floatingIPResourceName, "ip_address"),
					resource.TestCheckResourceAttrSet(floatingIPResourceName, "mac_address"),
					testAccCheckFloatingIP(floatingIPResourceName, firstServerResourceName),
				),
			},
			{
				Config: testUpcloudFloatingIPAddressCreateWithServerConfig([]string{"my_first_server", "my_second_server"}, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(floatingIPResourceName, "ip_address"),
					resource.TestCheckResourceAttrSet(floatingIPResourceName, "mac_address"),
					testAccCheckFloatingIP(floatingIPResourceName, secondServerResourceName),
				),
			},
		},
	})
}

func testAccCheckFloatingIP(floatingIPResourceName, serverResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		floatingIP := s.RootModule().Resources[floatingIPResourceName]
		server := s.RootModule().Resources[serverResourceName]

		if floatingIP == nil {
			return fmt.Errorf("Floating IP resource %s not found", floatingIPResourceName)
		}

		if server == nil {
			return fmt.Errorf("Server resource %s not found", serverResourceName)
		}

		serverNetworkMACAddress := server.Primary.Attributes["network_interface.0.mac_address"]
		floatingIPMACAddress := floatingIP.Primary.Attributes["mac_address"]

		if serverNetworkMACAddress != floatingIPMACAddress {
			return fmt.Errorf("server network MAC address and floating IP MAC address do not match: %s != %s", serverNetworkMACAddress, floatingIPMACAddress)
		}

		return nil
	}
}

func TestAccUpcloudFloatingIPAddress_import(t *testing.T) {
	var providers []*schema.Provider

	resourceName := "upcloud_floating_ip_address.my_floating_ip"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckFloatingIPAddressDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFloatingIPAddressBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "ip_address"),
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

func testAccCheckFloatingIPAddressDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "upcloud_floating_ip_address" {
			continue
		}

		client := testAccProvider.Meta().(*service.Service)
		addresses, err := client.GetIPAddresses()
		if err != nil {
			return fmt.Errorf("[WARN] Error listing Floating IP Addresses when deleting upcloud floating IP Address (%s): %s", rs.Primary.ID, err)
		}

		for _, IPAddress := range addresses.IPAddresses {
			if IPAddress.Address == rs.Primary.ID {
				return fmt.Errorf("[WARN] Tried deleting Floating IP (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testUpcloudFloatingIPAddressBasicConfig() string {
	return fmt.Sprintf(`
		resource "upcloud_floating_ip_address" "my_floating_ip" {
			zone = "fi-hel1"
		}
`)
}

func testUpcloudFloatingIPAddressCreateWithServerConfig(serverNames []string, assignedServerIndex int) string {

	config := strings.Builder{}

	for _, serverName := range serverNames {
		config.WriteString(fmt.Sprintf(`
		resource "upcloud_server" "%s" {
  			zone     = "fi-hel1"
  			hostname = "mydebian.example.com"
  			plan     = "1xCPU-2GB"

  			storage_devices {
    			action = "create"
    			size   = 10
    			tier   = "maxiops"
  			}

  			network_interface {
    			type = "public"
  			}
		}
	`, serverName))
	}

	config.WriteString(fmt.Sprintf(`
		resource "upcloud_floating_ip_address" "my_floating_ip" {
  			mac_address = upcloud_server.%s.network_interface[0].mac_address
		}
	`, serverNames[assignedServerIndex]))

	return config.String()
}
