package upcloud

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestUpcloudServer_basic(t *testing.T) {
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
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

			storage_devices {
					action = "create"
					size   = 10
					tier   = "maxiops"
			}

			network_interface {
				type = "utility"
			}
		}
	`)
}

func TestUpcloudServer_changePlan(t *testing.T) {
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
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

func TestUpcloudServer_networkInterface(t *testing.T) {
	var providers []*schema.Provider

	var serverID string

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccServerNetworkInterfaceConfig(
					networkInterface{
						niType: "utility",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "plan", "2xCPU-4GB"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "network_interface.#", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.0.type",
						"utility"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.0.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.ip_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.network"),
					testAccGetServerID("upcloud_server.my-server", &serverID),
				),
			},
			{
				Config: testAccServerNetworkInterfaceConfig(
					networkInterface{
						niType: "utility",
					},
					networkInterface{
						niType:  "private",
						network: true,
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "plan", "2xCPU-4GB"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "network_interface.#", "2"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.0.type",
						"utility"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.0.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.ip_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.network"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.1.type",
						"private"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.1.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.1.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.1.ip_address"),
					testAccCheckServerIDNotEqual("upcloud_server.my-server", serverID),
					testAccCheckNetwork("upcloud_server.my-server", 1, "upcloud_network.test_network_1"),
					testAccGetServerID("upcloud_server.my-server", &serverID),
				),
			},
			{
				Config: testAccServerNetworkInterfaceConfig(
					networkInterface{
						niType: "utility",
					},
					networkInterface{
						niType:     "private",
						network:    true,
						newNetwork: true,
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "plan", "2xCPU-4GB"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "network_interface.#", "2"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.0.type",
						"utility"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.0.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.ip_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.network"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.1.type",
						"private"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.1.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.1.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.1.ip_address"),
					testAccCheckServerIDNotEqual("upcloud_server.my-server", serverID),
					testAccCheckNetwork("upcloud_server.my-server", 1, "upcloud_network.test_network_11"),
				),
			},
		},
	})
}

func testAccGetServerID(resourceName string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		*id = s.RootModule().Resources[resourceName].Primary.ID

		return nil
	}
}

func testAccCheckServerIDNotEqual(resourceName string, id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		newID := s.RootModule().Resources[resourceName].Primary.ID
		if newID == id {
			return fmt.Errorf("new server ID unexpectedly equals old ID: %s == %s", newID, id)
		}

		return nil
	}
}

func testAccCheckNetwork(resourceName string, niIdx int, networkResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		server := s.RootModule().Resources[resourceName]
		network := s.RootModule().Resources[networkResourceName]
		if network == nil {
			return fmt.Errorf("network resource %s not found", networkResourceName)
		}

		serverNetworkID := server.Primary.Attributes[fmt.Sprintf("network_interface.%d.network", niIdx)]
		networkID := network.Primary.ID

		if serverNetworkID != networkID {
			return fmt.Errorf("server network ID and network ID do not match: %s != %s", serverNetworkID, networkID)
		}

		cidrRange := network.Primary.Attributes["ip_network.0.address"]
		serverIPStr := server.Primary.Attributes[fmt.Sprintf("network_interface.%d.ip_address", niIdx)]

		_, ipNet, err := net.ParseCIDR(cidrRange)
		if err != nil {
			return err
		}

		serverIP := net.ParseIP(serverIPStr)
		if !ipNet.Contains(serverIP) {
			return fmt.Errorf("server IP address is not in networks IP range: %s not in %s", serverIPStr, cidrRange)
		}

		return nil
	}
}

const testAccServerConfigWithSmallServerPlan = `
resource "upcloud_server" "my-server" {
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
`

const testAccPlanConfigUpdateServerPlan = `
resource "upcloud_server" "my-server" {
			zone     = "fi-hel1"
			hostname = "debian.example.com"
			plan     = "2xCPU-4GB"

			storage_devices {
					action = "create"
					size   = 10
					tier   = "maxiops"
			}

			network_interface {
				type = "utility"
			}
		}
`

type networkInterface struct {
	niType     string
	network    bool
	newNetwork bool
}

func testAccServerNetworkInterfaceConfig(nis ...networkInterface) string {
	var builder strings.Builder

	builder.WriteString(`
		resource "upcloud_server" "my-server" {
			zone     = "fi-hel1"
			hostname = "debian.example.com"
			plan     = "2xCPU-4GB"

			storage_devices {
					action = "create"
					size   = 10
					tier   = "maxiops"
			}
	`)

	for i, ni := range nis {
		builder.WriteString(fmt.Sprintf(`
				network_interface {
					type = "%s"
		`, ni.niType))

		if ni.network && !ni.newNetwork {
			builder.WriteString(fmt.Sprintf(`
						network = upcloud_network.test_network_%d.id
			`, i))
		} else if ni.newNetwork {
			builder.WriteString(fmt.Sprintf(`
						network = upcloud_network.test_network_%d.id
			`, 10+i))
		}
		builder.WriteString(`
				}
		`)
	}

	builder.WriteString(`
		}
	`)

	for i, ni := range nis {
		if ni.network {
			builder.WriteString(fmt.Sprintf(`
				resource "upcloud_network" "test_network_%d" {
					name = "test_network_%d"
					zone = "fi-hel1"

					ip_network {
						address = "10.0.%d.0/24"
						dhcp = true
						dhcp_default_route = false
						family = "IPv4"
						gateway = "10.0.%d.1"
					}
				}
			`, i, i, 14+i, 14+i))
		}

		if ni.newNetwork {
			builder.WriteString(fmt.Sprintf(`
				resource "upcloud_network" "test_network_%d" {
					name = "test_network_%d"
					zone = "fi-hel1"

					ip_network {
						address = "10.0.%d.0/24"
						dhcp = true
						dhcp_default_route = false
						family = "IPv4"
						gateway = "10.0.%d.1"
					}
				}
			`, 10+i, 10+i, 24+i, 24+i))
		}
	}

	return builder.String()
}
