package upcloud

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccUpCloudRouter(t *testing.T) {
	var providers []*schema.Provider

	var router upcloud.Router
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.my_example_router", &router),
					testAccCheckUpCloudRouterAttributes(&router, name),
				),
			},
		},
	})
}

func TestAccUpCloudRouter_update(t *testing.T) {
	var providers []*schema.Provider

	var router upcloud.Router
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	updateName := fmt.Sprintf("tf-test-update-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.my_example_router", &router),
					testAccCheckUpCloudRouterAttributes(&router, name),
				),
			},
			{
				Config: testAccRouterConfig(updateName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.my_example_router", &router),
					testAccCheckUpCloudRouterAttributes(&router, updateName),
				),
			},
		},
	})
}

func TestAccUpCloudRouter_import(t *testing.T) {
	var providers []*schema.Provider

	var router upcloud.Router
	name := fmt.Sprintf("tf-test-import-%s", acctest.RandString(10))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.my_example_router", &router),
				),
			},
			{
				ResourceName:      "upcloud_router.my_example_router",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUpCloudRouter_detach(t *testing.T) {
	var providers []*schema.Provider
	var router upcloud.Router
	var network upcloud.Network
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckRouterNetworkDestroy,
		Steps: []resource.TestStep{
			{
				// first create network and router attached
				Config: testAccRouterNetworkConfig("testrouter", "testnetwork", true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.terraform_test_router", &router),
					testAccCheckNetworkExists("upcloud_network.terraform_test_network", &network),
					testAccRouterAttachedNetworksCount(&router, 1),
					// make sure network and router are attached to each other
					testAccNetworkRouterAttached(&network, &router),
				),
			},
			{
				// and then change them to detached
				Config: testAccRouterNetworkConfig("testrouter", "testnetwork", true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.terraform_test_router", &router),
					testAccCheckNetworkExists("upcloud_network.terraform_test_network", &network),
					testAccRouterAttachedNetworksCount(&router, 0),
					// make sure network and router are NOT attached to each other
					testAccNetworkRouterNotAttached(&network, &router),
				),
			},
		},
	})
}

func TestAccUpCloudRouter_attachedDelete(t *testing.T) {
	var providers []*schema.Provider
	var router upcloud.Router
	var network upcloud.Network
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckRouterNetworkDestroy,
		Steps: []resource.TestStep{
			{
				// first create network and router attached
				Config: testAccRouterNetworkConfig("testrouter", "testnetwork", true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.terraform_test_router", &router),
					testAccCheckNetworkExists("upcloud_network.terraform_test_network", &network),
					testAccRouterAttachedNetworksCount(&router, 1),
					// make sure network and router are attached to each other
					testAccNetworkRouterAttached(&network, &router),
				),
			},
			{
				// and then try to delete the router
				Config: testAccRouterNetworkConfig("testrouter", "testnetwork", false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterDoesntExist("upcloud_router.terraform_test_router", &router),
					testAccCheckNetworkExists("upcloud_network.terraform_test_network", &network),
					// make sure network has no attachment anymore
					testAccNetworkNoRouterAttachment(&network),
				),
			},
		},
	})
}

func testAccCheckRouterExists(resourceName string, router *upcloud.Router) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Look for the full resource name and error if not found
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// The provider has not set the ID for the resource
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Router ID is set")
		}

		// Use the API SDK to locate the remote resource.
		client := testAccProvider.Meta().(*service.Service)
		latest, err := client.GetRouterDetails(&request.GetRouterDetailsRequest{
			UUID: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		// Update the reference the remote located router
		*router = *latest

		return nil
	}
}

func testAccCheckRouterDoesntExist(resourceName string, router *upcloud.Router) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Look for the full resource name in root module (internal state)
		if _, ok := s.RootModule().Resources[resourceName]; ok {
			return fmt.Errorf("router %s still exists in internal state", resourceName)
		}

		// Use the API SDK to locate the remote resource.
		client := testAccProvider.Meta().(*service.Service)
		_, err := client.GetRouterDetails(&request.GetRouterDetailsRequest{
			UUID: router.UUID,
		})

		if err == nil {
			return fmt.Errorf("router UUID %s still exists in remote", router.UUID)
		}

		return nil
	}
}

func testAccCheckNetworkExists(resourceName string, network *upcloud.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Look for the full resource name and error if not found
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// The provider has not set the ID for the resource
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network ID is set")
		}

		// Use the API SDK to locate the remote resource.
		client := testAccProvider.Meta().(*service.Service)
		latest, err := client.GetNetworkDetails(&request.GetNetworkDetailsRequest{
			UUID: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		// Update the reference the remote located network
		*network = *latest

		return nil
	}
}

func testAccCheckUpCloudRouterAttributes(router *upcloud.Router, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Confirm the remote router has the following attributes
		if router.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, router.Name)
		}

		return nil
	}
}

func testAccCheckRouterDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "upcloud_router" {
			continue
		}

		client := testAccProvider.Meta().(*service.Service)
		routers, err := client.GetRouters()
		if err != nil {
			return fmt.Errorf("[WARN] Error listing routers when deleting upcloud router (%s): %s", rs.Primary.ID, err)
		}

		for _, router := range routers.Routers {
			if router.UUID == rs.Primary.ID {
				// service still found
				return fmt.Errorf("[WARN] Tried deleting Router (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckRouterNetworkDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*service.Service)
	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "upcloud_router":
			routers, err := client.GetRouters()
			if err != nil {
				return fmt.Errorf("[WARN] Error listing routers when deleting upcloud router (%s): %s", rs.Primary.ID, err)
			}

			for _, router := range routers.Routers {
				if router.UUID == rs.Primary.ID {
					// service still found
					return fmt.Errorf("[WARN] Tried deleting Router (%s), but was still found", rs.Primary.ID)
				}
			}
		case "upcloud_network":
			networks, err := client.GetNetworks()
			if err != nil {
				return fmt.Errorf("[WARN] Error listing networks when deleting upcloud network (%s): %s", rs.Primary.ID, err)
			}

			for _, network := range networks.Networks {
				if network.UUID == rs.Primary.ID {
					// service still found
					return fmt.Errorf("[WARN] Tried deleting network (%s), but was still found", rs.Primary.ID)
				}
			}
		}
	}
	return nil
}

func testAccRouterConfig(name string) string {
	return fmt.Sprintf(`
resource "upcloud_router" "my_example_router" {
  name = "%s"
}`, name)
}

func testAccRouterNetworkConfig(routerName, networkName string, includeRouter, routerAttached bool) string {
	routerAttachment := ""
	if routerAttached {
		routerAttachment = "router = upcloud_router.terraform_test_router.id"
	}
	routerDefinition := ""
	if includeRouter {
		routerDefinition = fmt.Sprintf(`resource "upcloud_router" "terraform_test_router" {
  name = "%s"
}`, routerName)
	}

	return fmt.Sprintf(`
%s 

resource "upcloud_network" "terraform_test_network" {
  name = "%s"
  zone = "fi-hel1"

  %s

  ip_network {
    address            = "10.0.0.0/24"
    dhcp               = true
    family  = "IPv4"
  }
}
`, routerDefinition, networkName, routerAttachment)
}

func testAccNetworkRouterAttached(network *upcloud.Network, router *upcloud.Router) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if network.Router != router.UUID {
			return fmt.Errorf("network does not have the correct router attached, expected %s, got %s", router.UUID, network.Router)
		}
		found := false
		for _, attached := range router.AttachedNetworks {
			if attached.NetworkUUID == network.UUID {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("router does not have the correct network attached, expected %s, attached %s", network.UUID, router.AttachedNetworks)
		}
		return nil
	}
}

func testAccNetworkRouterNotAttached(network *upcloud.Network, router *upcloud.Router) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if network.Router == router.UUID {
			return fmt.Errorf("network %s still has the router %s attached", network.UUID, network.Router)
		}
		found := false
		for _, attached := range router.AttachedNetworks {
			if attached.NetworkUUID == network.UUID {
				found = true
			}
		}
		if found {
			return fmt.Errorf("router still has network %s attached attached %s", network.UUID, router.AttachedNetworks)
		}
		return nil
	}
}

func testAccNetworkNoRouterAttachment(network *upcloud.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if network.Router != "" {
			return fmt.Errorf("network %s still has the router %s attached", network.UUID, network.Router)
		}
		return nil
	}
}

func testAccRouterAttachedNetworksCount(router *upcloud.Router, expectedAttachedNetworksCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(router.AttachedNetworks) != expectedAttachedNetworksCount {
			return fmt.Errorf("router does not have the correct number of networks, expected %d, got %d", expectedAttachedNetworksCount, len(router.AttachedNetworks))
		}
		return nil
	}
}
