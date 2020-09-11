package upcloud

import (
	"fmt"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

func testAccRouterConfig(name string) string {
	return fmt.Sprintf(`
resource "upcloud_router" "my_example_router" {
  name = "%s"
}`, name)
}
