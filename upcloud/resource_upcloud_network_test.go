package upcloud

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUpCloudNetwork_basic(t *testing.T) {
	var providers []*schema.Provider

	netName := fmt.Sprintf("test_network_%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("10.0.%d.0/24", subnet)
	gateway := fmt.Sprintf("10.0.%d.1", subnet)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfig(netName, "fi-hel1", cidr, gateway, true, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccNetworkExists("upcloud_network.test_network"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "name", netName),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "zone", "fi-hel1"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp", "true"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_default_route", "false"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.address", cidr),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.gateway", gateway),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.family", "IPv4"),
				),
			},
		},
	})
}

func TestAccUpCloudNetwork_basicUpdate(t *testing.T) {
	var providers []*schema.Provider

	netName := fmt.Sprintf("test_network_%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("10.0.%d.0/24", subnet)
	gateway := fmt.Sprintf("10.0.%d.1", subnet)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfig(netName, "fi-hel1", cidr, gateway, true, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccNetworkExists("upcloud_network.test_network"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "name", netName),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "zone", "fi-hel1"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp", "true"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_default_route", "false"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.address", cidr),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.gateway", gateway),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.family", "IPv4"),
				),
			},
			{
				Config: testAccNetworkConfig(netName+"_1", "fi-hel1", cidr, gateway, true, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccNetworkExists("upcloud_network.test_network"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "name", netName+"_1"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "zone", "fi-hel1"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp", "true"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.dhcp_default_route", "false"),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.address", cidr),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.gateway", gateway),
					resource.TestCheckResourceAttr("upcloud_network.test_network", "ip_network.0.family", "IPv4"),
				),
			},
		},
	})
}

func TestAccUpCloudNetwork_withRouter(t *testing.T) {
	var providers []*schema.Provider

	netName := fmt.Sprintf("test_network_%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("10.0.%d.0/24", subnet)
	gateway := fmt.Sprintf("10.0.%d.1", subnet)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfig(netName, "fi-hel1", cidr, gateway, true, false, true),
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

func TestAccUpCloudNetwork_amendWithRouter(t *testing.T) {
	var providers []*schema.Provider

	netName := fmt.Sprintf("test_network_%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("10.0.%d.0/24", subnet)
	gateway := fmt.Sprintf("10.0.%d.1", subnet)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfig(netName, "fi-hel1", cidr, gateway, true, false, false),
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
				Config: testAccNetworkConfig(netName, "fi-hel1", cidr, gateway, true, false, true),
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
	var providers []*schema.Provider

	netName := fmt.Sprintf("test_network_%s", acctest.RandString(5))
	subnet := acctest.RandIntRange(0, 250)
	cidr := fmt.Sprintf("10.0.%d.0/24", subnet)
	gateway := fmt.Sprintf("10.0.%d.1", subnet)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkConfigWithFamily(netName, "fi-hel1", cidr, gateway, "rubbish", true, false, false),
				ExpectError: regexp.MustCompile(`'family' has incorrect value`),
			},
		},
	})
}

func testAccNetworkConfig(
	name string,
	zone string,
	address string,
	gateway string,
	dhcp bool,
	dhcpDefaultRoute bool,
	router bool) string {

	return testAccNetworkConfigWithFamily(name, zone, address, gateway, "IPv4", dhcp, dhcpDefaultRoute, router)
}

func testAccNetworkConfigWithFamily(
	name string,
	zone string,
	address string,
	gateway string,
	family string,
	dhcp bool,
	dhcpDefaultRoute bool,
	router bool) string {

	config := strings.Builder{}

	config.WriteString(fmt.Sprintf(`
	  resource "upcloud_network" "test_network" {
		name = "%s"
		zone = "%s"
	`, name, zone))

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
		  family  = "%s"
		  gateway = "%s"
		}
	  }
	`, address, dhcp, dhcpDefaultRoute, family, gateway))

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
