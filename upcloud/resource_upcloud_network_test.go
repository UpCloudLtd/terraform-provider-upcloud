package upcloud

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUpCloudNetwork_basic(t *testing.T) {
	netName := fmt.Sprintf("test_network_%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("10.0.%d.0/24", subnet)
	gateway := fmt.Sprintf("10.0.%d.1", subnet)

	config := testAccNetworkConfig(
		netName,
		"fi-hel1",
		cidr,
		gateway,
		true,
		false,
		false,
		[]string{"10.0.0.2", "10.0.0.3"},
		[]string{"192.168.0.0/24", "192.168.100.0/32"},
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccNetworkExists("upcloud_network.test_network"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "name", netName),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "zone", "fi-hel1"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp", "true"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_default_route", "false"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.address", cidr),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.gateway", gateway),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.family", "IPv4"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_dns.#", "2"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_routes.#", "2"),
				),
			},
			{
				Config:            config,
				ResourceName:      "upcloud_network.test_network",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUpCloudNetwork_basicUpdate(t *testing.T) {
	netName := fmt.Sprintf("test_network_%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("10.0.%d.0/24", subnet)
	gateway := fmt.Sprintf("10.0.%d.1", subnet)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfig(netName, "fi-hel1", cidr, gateway, true, false, false, []string{"10.0.0.2"}, []string{"192.168.0.0/24"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccNetworkExists("upcloud_network.test_network"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "name", netName),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "zone", "fi-hel1"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp", "true"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_default_route", "false"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.address", cidr),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.gateway", gateway),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.family", "IPv4"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_dns.0", "10.0.0.2"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_routes.0", "192.168.0.0/24"),
				),
			},
			{
				Config: testAccNetworkConfig(netName+"_1", "fi-hel1", cidr, gateway, true, false, false, []string{"10.0.0.3"}, []string{"192.168.100.0/24"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccNetworkExists("upcloud_network.test_network"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "name", netName+"_1"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "zone", "fi-hel1"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp", "true"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_default_route", "false"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.address", cidr),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.gateway", gateway),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.family", "IPv4"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_dns.0", "10.0.0.3"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_routes.0", "192.168.100.0/24"),
				),
			},
		},
	})
}

func TestAccUpCloudNetwork_withRouter(t *testing.T) {
	netName := fmt.Sprintf("test_network_%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("10.0.%d.0/24", subnet)
	gateway := fmt.Sprintf("10.0.%d.1", subnet)

	config := testAccNetworkConfig(netName, "fi-hel1", cidr, gateway, true, false, true, nil, nil)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccNetworkExists("upcloud_network.test_network"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "name", netName),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "zone", "fi-hel1"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp", "true"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_default_route", "false"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.address", cidr),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.gateway", gateway),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.family", "IPv4"),
					testAccNetworkRouterIsSet("upcloud_network.test_network", "upcloud_router.test_network_router"),
				),
			},
			{
				Config:            config,
				ResourceName:      "upcloud_network.test_network",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUpCloudNetwork_amendWithRouter(t *testing.T) {
	netName := fmt.Sprintf("test_network_%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("10.0.%d.0/24", subnet)
	gateway := fmt.Sprintf("10.0.%d.1", subnet)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfig(netName, "fi-hel1", cidr, gateway, true, false, false, nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccNetworkExists("upcloud_network.test_network"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "name", netName),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "zone", "fi-hel1"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp", "true"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_default_route", "false"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.address", cidr),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.gateway", gateway),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.family", "IPv4"),
					resource.TestCheckNoResourceAttr("upcloud_network.test_network", "router"),
				),
			},
			{
				Config: testAccNetworkConfig(netName, "fi-hel1", cidr, gateway, true, false, true, nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccNetworkExists("upcloud_network.test_network"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "name", netName),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "zone", "fi-hel1"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp", "true"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_default_route", "false"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.address", cidr),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.gateway", gateway),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.family", "IPv4"),
					testAccNetworkRouterIsSet("upcloud_network.test_network", "upcloud_router.test_network_router"),
				),
			},
		},
	})
}

func TestAccUpCloudNetwork_FamilyValidation(t *testing.T) {
	netName := fmt.Sprintf("test_network_%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("10.0.%d.0/24", subnet)
	gateway := fmt.Sprintf("10.0.%d.1", subnet)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkConfigWithFamily(netName, "fi-hel1", cidr, gateway, "rubbish", true, false, false, nil, nil, nil),
				ExpectError: regexp.MustCompile(`family value must be one of: \["IPv4" "IPv6"\]`),
			},
		},
	})
}

func TestAccUpCloudNetwork_labels(t *testing.T) {
	netName := fmt.Sprintf("test_network_labels_%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("10.0.%d.0/24", subnet)

	config1 := testAccNetworkLabelsConfig(
		netName,
		"fi-hel1",
		cidr,
		map[string]string{
			"key":    "value",
			"test":   "tf-acc-test",
			"animal": "cow",
		},
	)
	config2 := testAccNetworkLabelsConfig(
		netName,
		"fi-hel1",
		cidr,
		map[string]string{
			"key":  "",
			"test": "tf-acc-test",
		},
	)
	config3 := testAccNetworkLabelsConfig(
		netName,
		"fi-hel1",
		cidr,
		nil,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccNetworkExists("upcloud_network.test_network"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "labels.%", "3"),
				),
			},
			{
				Config:            config1,
				ResourceName:      "upcloud_network.test_network",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config2,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccNetworkExists("upcloud_network.test_network"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "labels.%", "2"),
				),
			},
			{
				Config: config3,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccNetworkExists("upcloud_network.test_network"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "labels.%", "0"),
				),
			},
		},
	})
}

func TestAccUpcloudNetwork_EffectiveRoutes(t *testing.T) {
	configStep1 := utils.ReadTestDataFile(t, "testdata/upcloud_network/network_cfg1.tf")

	prefix := "tf-acc-test-network-"
	netName := fmt.Sprintf("file-storage-net-%s", acctest.RandString(5))
	routerName := fmt.Sprintf("network-router-%s", acctest.RandString(5))
	randOctet := acctest.RandIntRange(10, 250)
	networkCIDR := fmt.Sprintf("10.%d.0.0/24", randOctet)
	gatewayIP := fmt.Sprintf("10.%d.0.1", randOctet)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configStep1,
				ConfigVariables: map[string]config.Variable{
					"prefix":       config.StringVariable(prefix),
					"net-name":     config.StringVariable(netName),
					"router-name":  config.StringVariable(routerName),
					"network-cidr": config.StringVariable(networkCIDR),
					"gateway-ip":   config.StringVariable(gatewayIP),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_network.test", "effective_routes.#"),
					resource.TestCheckTypeSetElemNestedAttrs("upcloud_network.test", "effective_routes.*", map[string]string{
						"route": "192.168.0.0/24",
					}),
					resource.TestCheckResourceAttrSet("upcloud_network.test", "ip_network.0.dhcp_effective_routes.#"),
					resource.TestCheckTypeSetElemNestedAttrs("upcloud_network.test", "ip_network.0.dhcp_effective_routes.*", map[string]string{
						"route":   "192.168.0.0/24",
						"nexthop": gatewayIP,
					}),
				),
			},
			{
				// Import verification ensures state matches real API
				ResourceName:      "upcloud_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUpcloudNetwork_DHCPRoutesConfiguration(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: Create with DHCP routes auto-population enabled, no filters
				Config: `
					resource "upcloud_network" "test" {
						name = "tf-acc-test-dhcp-routes"
						zone = "fi-hel1"

						ip_network {
							address            = "10.20.0.0/24"
							dhcp               = true
							dhcp_default_route = true
							family             = "IPv4"
							gateway            = "10.20.0.1"

							dhcp_routes_configuration = {
								effective_routes_auto_population = {
									enabled = true
								}
							}
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_network.test", "ip_network.0.dhcp", "true"),
					resource.TestCheckResourceAttr("upcloud_network.test", "ip_network.0.dhcp_default_route", "true"),
					resource.TestCheckResourceAttr("upcloud_network.test", "ip_network.0.dhcp_routes_configuration.effective_routes_auto_population.enabled", "true"),
				),
			},
			{
				// Step 2: Update with filters
				Config: `
					resource "upcloud_network" "test" {
						name = "tf-acc-test-dhcp-routes"
						zone = "fi-hel1"

						ip_network {
							address            = "10.20.0.0/24"
							dhcp               = true
							dhcp_default_route = true
							family             = "IPv4"
							gateway            = "10.20.0.1"

							dhcp_routes_configuration = {
								effective_routes_auto_population = {
									enabled = true

									filter_by_destination = [
										"10.30.0.0/24",
										"172.16.0.0/22"
									]

									exclude_by_source = [
										"static-route",
									]

									filter_by_route_type = [
										"service",
									]
								}
							}
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_network.test", "ip_network.0.dhcp_routes_configuration.effective_routes_auto_population.enabled", "true"),
					resource.TestCheckTypeSetElemAttr("upcloud_network.test", "ip_network.0.dhcp_routes_configuration.effective_routes_auto_population.filter_by_destination.*", "10.30.0.0/24"),
					resource.TestCheckTypeSetElemAttr("upcloud_network.test", "ip_network.0.dhcp_routes_configuration.effective_routes_auto_population.filter_by_destination.*", "172.16.0.0/22"),
					resource.TestCheckTypeSetElemAttr("upcloud_network.test", "ip_network.0.dhcp_routes_configuration.effective_routes_auto_population.exclude_by_source.*", "static-route"),
					resource.TestCheckTypeSetElemAttr("upcloud_network.test", "ip_network.0.dhcp_routes_configuration.effective_routes_auto_population.filter_by_route_type.*", "service"),
				),
			},
			{
				// Step 3: Clear filters
				Config: `
					resource "upcloud_network" "test" {
						name = "tf-acc-test-dhcp-routes"
						zone = "fi-hel1"

						ip_network {
							address            = "10.20.0.0/24"
							dhcp               = true
							dhcp_default_route = true
							family             = "IPv4"
							gateway            = "10.20.0.1"

							dhcp_routes_configuration = {
								effective_routes_auto_population = {
									enabled = true

									// Explicitly clear all filters
									filter_by_destination = []
									exclude_by_source     = []
									filter_by_route_type  = []
								}
							}
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_network.test", "ip_network.0.dhcp_routes_configuration.effective_routes_auto_population.enabled", "true"),

					// Empty sets should have size 0
					resource.TestCheckResourceAttr("upcloud_network.test", "ip_network.0.dhcp_routes_configuration.effective_routes_auto_population.filter_by_destination.#", "0"),
					resource.TestCheckResourceAttr("upcloud_network.test", "ip_network.0.dhcp_routes_configuration.effective_routes_auto_population.exclude_by_source.#", "0"),
					resource.TestCheckResourceAttr("upcloud_network.test", "ip_network.0.dhcp_routes_configuration.effective_routes_auto_population.filter_by_route_type.#", "0"),
				),
			},
			{
				ResourceName:      "upcloud_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNetworkConfig(name string, zone string, address string, gateway string, dhcp bool, dhcpDefaultRoute bool, router bool, dhcpDNS []string, dhcpRoutes []string) string {
	return testAccNetworkConfigWithFamily(name, zone, address, gateway, "IPv4", dhcp, dhcpDefaultRoute, router, dhcpDNS, dhcpRoutes, nil)
}

func testAccNetworkLabelsConfig(name string, zone string, address string, labels map[string]string) string {
	return testAccNetworkConfigWithFamily(name, zone, address, "", "IPv4", true, true, false, nil, nil, labels)
}

func testAccNetworkConfigWithFamily(name string, zone string, address string, gateway string, family string, dhcp bool, dhcpDefaultRoute bool, router bool, dhcpDNS []string, dhcpRoutes []string, labels map[string]string) string {
	config := strings.Builder{}

	config.WriteString(fmt.Sprintf(`
	  resource "upcloud_network" "test_network" {
		name = "%s"
		zone = "%s"
	`, name, zone))

	if len(labels) > 0 {
		config.WriteString(`
		labels = {
		`)
		for k, v := range labels {
			config.WriteString(fmt.Sprintf(`
			"%s" = "%s"`, k, v))
		}
		config.WriteString(`
		}
		`)
	}

	if router {
		config.WriteString(`
		router = upcloud_router.test_network_router.id
		`)
	}

	config.WriteString(fmt.Sprintf(`
		ip_network {
		  address            = "%s"
		  dhcp               = "%t"
		  dhcp_default_route = "%t"
		  family             = "%s"
		
	`,
		address,
		dhcp,
		dhcpDefaultRoute,
		family))

	if gateway != "" {
		config.WriteString(fmt.Sprintf(`
		  gateway            = "%s"`, gateway))
	}

	if len(dhcpDNS) > 0 {
		config.WriteString(fmt.Sprintf(`
		  dhcp_dns			 = ["%s"]`, strings.Join(dhcpDNS, "\", \"")))
	}

	if len(dhcpRoutes) > 0 {
		config.WriteString(fmt.Sprintf(`
		  dhcp_routes		 = ["%s"]`, strings.Join(dhcpRoutes, "\", \"")))
	}

	config.WriteString(`
	    }
	  }
	`)

	if router {
		config.WriteString(fmt.Sprintf(`
		  resource "upcloud_router" "test_network_router" {
			  name = "%s_router"
		  }
		`, name))
	}

	return config.String()
}

func testAccNetworkExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Check that the expected resource exists
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		return nil
	}
}

func testAccNetworkRouterIsSet(netResourceName string, routerResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		network, ok := s.RootModule().Resources[netResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", netResourceName)
		}

		router, ok := s.RootModule().Resources[routerResourceName]
		if !ok {
			return fmt.Errorf("router not found: %s", routerResourceName)
		}

		routerID := router.Primary.ID

		netRouterID, ok := network.Primary.Attributes["router"]
		if !ok {
			return errors.New("network router attribute not found")
		}

		if netRouterID != routerID {
			return fmt.Errorf("network router ID does not match router ID: (%s != %s)", netRouterID, routerID)
		}

		return nil
	}
}
