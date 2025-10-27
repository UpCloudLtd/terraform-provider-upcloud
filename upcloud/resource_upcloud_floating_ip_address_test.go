package upcloud

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	zone                   = "fi-hel1"
	floatingIPResourceName = "upcloud_floating_ip_address.test"
)

func TestAccUpcloudFloatingIPAddress_basic(t *testing.T) {
	resourceName := floatingIPResourceName
	expectedZone := zone
	expectedFamily := "IPv4"
	expectedAccess := "public"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		CheckDestroy:             testAccCheckFloatingIPAddressDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFloatingIPAddressBasicConfig(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "zone", expectedZone),
					resource.TestCheckResourceAttr(resourceName, "family", expectedFamily),
					resource.TestCheckResourceAttr(resourceName, "access", expectedAccess),
					resource.TestCheckResourceAttrSet(resourceName, "ip_address"),
					// API default for release policy is "keep"
					resource.TestCheckResourceAttr(resourceName, "release_policy", "keep"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testUpcloudFloatingIPAddressBasicConfig("release"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "release_policy", "release"),
				),
			},
		},
	})
}

func TestAccUpcloudFloatingIPAddress_create_with_server(t *testing.T) {
	serverResourceName := "upcloud_server.test"
	expectedZone := zone
	expectedFamily := "IPv4"
	expectedAccess := "public"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFloatingIPAddressCreateWithServerConfig([]string{"test"}, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(floatingIPResourceName, "zone", expectedZone),
					resource.TestCheckResourceAttr(floatingIPResourceName, "family", expectedFamily),
					resource.TestCheckResourceAttr(floatingIPResourceName, "access", expectedAccess),
					resource.TestCheckResourceAttrSet(floatingIPResourceName, "ip_address"),
					resource.TestCheckResourceAttrSet(floatingIPResourceName, "mac_address"),
					testAccCheckFloatingIP(floatingIPResourceName, serverResourceName),
				),
			},
			{
				ResourceName:      floatingIPResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUpcloudFloatingIPAddress_switch_between_servers(t *testing.T) {
	firstServerResourceName := "upcloud_server.first"
	secondServerResourceName := "upcloud_server.second"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudFloatingIPAddressCreateWithServerConfig([]string{"first", "second"}, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(floatingIPResourceName, "ip_address"),
					resource.TestCheckResourceAttrSet(floatingIPResourceName, "mac_address"),
					testAccCheckFloatingIP(floatingIPResourceName, firstServerResourceName),
				),
			},
			{
				Config: testUpcloudFloatingIPAddressCreateWithServerConfig([]string{"first", "second"}, 1),
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

func testAccCheckFloatingIPAddressDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "upcloud_floating_ip_address" {
			continue
		}

		client := testAccProvider.Meta().(*service.Service)
		addresses, err := client.GetIPAddresses(context.Background())
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

func testUpcloudFloatingIPAddressBasicConfig(releasePolicy string) string {
	config := strings.Builder{}
	config.WriteString(`
		resource "upcloud_floating_ip_address" "test" {
			zone = "fi-hel1"
	`)

	if releasePolicy != "" {
		config.WriteString(fmt.Sprintf(`
			release_policy = "%s"
		`, releasePolicy))
	}

	config.WriteString(`
		}
	`)
	return config.String()
}

func testUpcloudFloatingIPAddressCreateWithServerConfig(serverNames []string, assignedServerIndex int) string {
	config := strings.Builder{}

	for _, serverName := range serverNames {
		config.WriteString(fmt.Sprintf(`
			resource "upcloud_server" "%s" {
				zone     = "fi-hel1"
				hostname = "tf-acc-test-floating-ip-vm"
				plan     = "1xCPU-2GB"
				metadata = true

				template {
					storage = "%s"
					size = 10
				}

				network_interface {
					type = "public"
				}
			}
		`, serverName, debianTemplateUUID))
	}

	config.WriteString(fmt.Sprintf(`
		resource "upcloud_floating_ip_address" "test" {
  			mac_address = upcloud_server.%s.network_interface[0].mac_address
		}
	`, serverNames[assignedServerIndex]))

	return config.String()
}
