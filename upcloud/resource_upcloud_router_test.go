package upcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUpCloudRouter(t *testing.T) {
	var router upcloud.Router
	name := fmt.Sprintf("tf-acc-test-router-%s", acctest.RandString(10))

	staticRoutes := []upcloud.StaticRoute{{Name: "test-route", Nexthop: "10.0.0.100", Route: "0.0.0.0/0"}}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		CheckDestroy:             testAccCheckRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfig(name, staticRoutes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.this", &router),
					testAccCheckUpCloudRouterAttributes(&router, name),
					resource.TestCheckTypeSetElemNestedAttrs("upcloud_router.this", "static_route.*", map[string]string{
						"name":    "test-route",
						"nexthop": "10.0.0.100",
						"route":   "0.0.0.0/0",
					}),
				),
			},
			{
				ResourceName:      "upcloud_router.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUpCloudRouter_update(t *testing.T) {
	var router upcloud.Router
	name := fmt.Sprintf("tf-acc-test-router-%s", acctest.RandString(10))
	updateName := fmt.Sprintf("tf-acc-test-router-update-%s", acctest.RandString(10))

	staticRoutes := []upcloud.StaticRoute{{Nexthop: "10.0.0.100", Route: "0.0.0.0/0"}}
	updateStaticRoutes := []upcloud.StaticRoute{{Name: "test-route-2", Nexthop: "10.0.0.101", Route: "0.0.0.0/0"}}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		CheckDestroy:             testAccCheckRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfig(name, staticRoutes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.this", &router),
					testAccCheckUpCloudRouterAttributes(&router, name),
					resource.TestCheckTypeSetElemNestedAttrs("upcloud_router.this", "static_route.*", map[string]string{
						"name":    "static-route-0",
						"nexthop": "10.0.0.100",
						"route":   "0.0.0.0/0",
					}),
				),
			},
			{
				Config: testAccRouterConfig(updateName, updateStaticRoutes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.this", &router),
					testAccCheckUpCloudRouterAttributes(&router, updateName),
					resource.TestCheckTypeSetElemNestedAttrs("upcloud_router.this", "static_route.*", map[string]string{
						"name":    "test-route-2",
						"nexthop": "10.0.0.101",
						"route":   "0.0.0.0/0",
					}),
				),
			},
		},
	})
}

func TestAccUpCloudRouter_detach(t *testing.T) {
	testDataStep1 := utils.ReadTestDataFile(t, "testdata/upcloud_router/detach_s1.tf")
	testDataStep2 := utils.ReadTestDataFile(t, "testdata/upcloud_router/detach_s2.tf")

	var router upcloud.Router
	var network upcloud.Network
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		CheckDestroy:             testAccCheckRouterNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDataStep1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.this", &router),
					testAccCheckNetworkExists("upcloud_network.this", &network),
					testAccRouterAttachedNetworksCount(&router, 1),
					// make sure network and router are attached to each other
					testAccNetworkRouterAttached(&network, &router),
				),
			},
			{
				Config: testDataStep2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.this", &router),
					testAccCheckNetworkExists("upcloud_network.this", &network),
					testAccRouterAttachedNetworksCount(&router, 0),
					// make sure network and router are NOT attached to each other
					testAccNetworkRouterNotAttached(&network, &router),
				),
			},
		},
	})
}

func TestAccUpCloudRouter_attachedDelete(t *testing.T) {
	testDataStep1 := utils.ReadTestDataFile(t, "testdata/upcloud_router/delete_attached_s1.tf")
	testDataStep2 := utils.ReadTestDataFile(t, "testdata/upcloud_router/delete_attached_s2.tf")

	var router upcloud.Router
	var network upcloud.Network
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		CheckDestroy:             testAccCheckRouterNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDataStep1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterExists("upcloud_router.this", &router),
					testAccCheckNetworkExists("upcloud_network.this", &network),
					testAccRouterAttachedNetworksCount(&router, 1),
					// make sure network and router are attached to each other
					testAccNetworkRouterAttached(&network, &router),
				),
			},
			{
				Config: testDataStep2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouterDoesntExist("upcloud_router.this", &router),
					testAccCheckNetworkExists("upcloud_network.this", &network),
					// make sure network has no attachment anymore
					testAccNetworkNoRouterAttachment(&network),
				),
			},
		},
	})
}

func TestAccUpCloudRouter_staticRoutes(t *testing.T) {
	testDataStep1 := utils.ReadTestDataFile(t, "testdata/upcloud_router/static_routes_s1.tf")
	testDataStep2 := utils.ReadTestDataFile(t, "testdata/upcloud_router/static_routes_s2.tf")

	router := "upcloud_router.this"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(router, "static_route.#", "0"),
				),
			},
			{
				Config: testDataStep2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(router, "static_route.#", "1"),
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
		client := TestAccProvider.Meta().(*service.Service)
		latest, err := client.GetRouterDetails(context.Background(), &request.GetRouterDetailsRequest{
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
		client := TestAccProvider.Meta().(*service.Service)
		_, err := client.GetRouterDetails(context.Background(), &request.GetRouterDetailsRequest{
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
		client := TestAccProvider.Meta().(*service.Service)
		latest, err := client.GetNetworkDetails(context.Background(), &request.GetNetworkDetailsRequest{
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
	return func(_ *terraform.State) error {
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

		client := TestAccProvider.Meta().(*service.Service)
		routers, err := client.GetRouters(context.Background())
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
	client := TestAccProvider.Meta().(*service.Service)
	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "upcloud_router":
			routers, err := client.GetRouters(context.Background())
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
			networks, err := client.GetNetworks(context.Background())
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

func testAccRouterConfig(name string, staticRoutes []upcloud.StaticRoute) string {
	s := fmt.Sprintf(`
resource "upcloud_router" "this" {
  name = "%s"
`, name)

	if len(staticRoutes) > 0 {
		for _, staticRoute := range staticRoutes {
			s = s + fmt.Sprintf(`
  static_route {
    nexthop = "%s"
    route   = "%s"
`, staticRoute.Nexthop, staticRoute.Route)

			if len(staticRoute.Name) > 0 {
				s = s + fmt.Sprintf(`
    name    = "%s"
`, staticRoute.Name)
			}
		}
		s = s + `
  }`
	}
	s = s + `
}
`
	return s
}

func testAccNetworkRouterAttached(network *upcloud.Network, router *upcloud.Router) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
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
	return func(_ *terraform.State) error {
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
	return func(_ *terraform.State) error {
		if network.Router != "" {
			return fmt.Errorf("network %s still has the router %s attached", network.UUID, network.Router)
		}
		return nil
	}
}

func testAccRouterAttachedNetworksCount(router *upcloud.Router, expectedAttachedNetworksCount int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if len(router.AttachedNetworks) != expectedAttachedNetworksCount {
			return fmt.Errorf("router does not have the correct number of networks, expected %d, got %d", expectedAttachedNetworksCount, len(router.AttachedNetworks))
		}
		return nil
	}
}
