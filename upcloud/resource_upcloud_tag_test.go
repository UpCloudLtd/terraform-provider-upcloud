package upcloud

import (
	"fmt"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudTag_basic(t *testing.T) {
	var providers []*schema.Provider

	expectedNames := []string{"dev", "test"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudTagInstanceConfig(expectedNames),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_tag.my_tag_dev", "name"),
					resource.TestCheckResourceAttr("upcloud_tag.my_tag_dev", "name", expectedNames[0]),
					resource.TestCheckResourceAttrSet("upcloud_tag.my_tag_test", "name"),
					resource.TestCheckResourceAttr("upcloud_tag.my_tag_test", "name", expectedNames[1]),
				),
			},
		},
	})
}

func TestAccUpCloudTag_import(t *testing.T) {
	var providers []*schema.Provider
	var tags upcloud.Tags

	expectedNames := []string{"dev"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudTagInstanceConfig(expectedNames),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagsExists("upcloud_tag.my_tag_dev", &tags),
				),
			},
			{
				ResourceName:      "upcloud_tag.my_tag_dev",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

}

func testAccCheckTagsExists(resourceName string, tags *upcloud.Tags) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Look for the full resource name and error if not found
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// The provider has not set the ID for the resource
		if rs.Primary.ID == "" {
			return fmt.Errorf("No tag ID is set")
		}

		// Use the API SDK to locate the remote resource.
		client := testAccProvider.Meta().(*service.Service)
		latest, err := client.GetTags()

		if err != nil {
			return err
		}

		// Update the reference the remote located tags
		*tags = *latest

		return nil
	}
}

func testAccCheckTagDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "upcloud_tag" {
			continue
		}

		client := testAccProvider.Meta().(*service.Service)
		tags, err := client.GetTags()
		if err != nil {
			return fmt.Errorf("[WARN] Error listing tags when deleting upcloud tag (%s): %s", rs.Primary.ID, err)
		}

		for _, tag := range tags.Tags {
			if tag.Name == rs.Primary.ID {
				return fmt.Errorf("[WARN] Tried deleting tag (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testUpcloudTagInstanceConfig(names []string) string {

	config := strings.Builder{}

	for _, name := range names {
		config.WriteString(fmt.Sprintf(`
		resource "upcloud_tag" "my_tag_%s" {
  			name = "%s"
  			description = "Represents the %s environment"
		}`, name, name, name))
	}

	return config.String()
}
